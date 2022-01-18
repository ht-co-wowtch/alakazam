package member

import (
	"database/sql"

	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"gitlab.com/ht-co/cpw/micro/log"
	"gitlab.com/ht-co/wowtch/live/alakazam/client"
	"gitlab.com/ht-co/wowtch/live/alakazam/errors"
	"gitlab.com/ht-co/wowtch/live/alakazam/models"
	"go.uber.org/zap"
)

const (
	RootMid       = 1
	RootUid       = "root"
	RootName      = "管理员"
	System        = "系统"
	LevelUp       = "升级"
	GeneralMember = "会员"
	Private       = "私讯"
	RoomAdmin     = "房管"
	Anchor        = "主播"
	Win           = "中奖"
)

type Chat interface {
	Get(uid string) (*models.Member, error)
	GetSession(uid string) (*models.Member, error)
	GetMessageSession(uid string, rid int) (*models.Member, error)
	GetByRoom(uid string, rid int) (*models.Member, error)
	Login(room models.Room, token, server string) (*models.Member, string, error)
	Logout(uid, key string) (bool, error)
	ChangeRoom(uid, key string, rid int) error
	Heartbeat(uid string) error
}

type Member struct {
	cli *client.Client
	db  models.Chat
	c   Cache
}

func New(db models.Chat, cache *redis.Client, cli *client.Client) *Member {
	return &Member{
		db:  db,
		cli: cli,
		c:   NewCache(cache),
	}
}

var (
	errInsertMember = errors.New("insert member")
)

func (m *Member) Login(room models.Room, token, server string) (*models.Member, string, error) {
	user, err := m.cli.Auth(token)
	if err != nil {
		return nil, "", err
	}

	u, err := m.GetByRoom(user.Uid, room.Id)

	if err == errors.ErrNoMember {
		lv, err := m.cli.Level(user.Uid)
		if err != nil {
			lv = 0
		}

		u = &models.Member{
			Uid:    user.Uid,
			Name:   user.Name,
			Gender: user.Gender,
			Type:   user.Type,
			Lv:     lv,
		}

		ok, err := m.db.CreateUser(u)
		if err != nil {
			return nil, "", err
		}
		if !ok {
			return nil, "", errInsertMember
		}
	} else if err != nil {
		return nil, "", err
	}

	// 會員等級
	lv, err := m.cli.Level(user.Uid)
	if err != nil {
		lv = 0
	}

	log.Infof("user lv", zap.Int("lv", user.Lv))
	log.Infof("u lv", zap.Int("lv", u.Lv))
	log.Infof("lv", zap.Int("lv", lv))
	
	u.Lv = lv
	log.Infof("u lv", zap.Int("lv", u.Lv))


	if u.Blockade() {
		return u, "", nil
	}

	if u.Name != user.Name || u.Gender != user.Gender || u.Lv != user.Lv {
		u.Name = user.Name
		u.Gender = user.Gender
		if ok, err := m.db.UpdateUser(u); err != nil || !ok {
			log.Error("update user", zap.String("uid", user.Uid), zap.Bool("action", ok), zap.Error(err))
		}
	}

	key := uuid.New().String()

	if err = m.c.login(u, room.Id, key); err != nil {
		return nil, "", err
	} else {
		log.Info(
			"[logic/member/member.go connected]",
			zap.String("key", key),
			zap.String("uid", u.Uid),
			zap.Int("room_id", room.Id),
			zap.String("server", server),
		)
	}

	return u, key, nil
}

// Logout
// 登出聊天室
func (m *Member) Logout(uid, key string) (bool, error) {
	// 從快取中刪除聊天室連線紀錄
	return m.c.logout(uid, key)
}

// ChangeRoom
// 切換聊天室房間
func (m *Member) ChangeRoom(uid, key string, rid int) error {
	return m.c.setWs(uid, key, rid)
}

func (m *Member) SetManage(uid string, rid int, set bool) error {
	member, err := m.GetByRoom(uid, rid)
	if err != nil {
		return err
	}

	member.Permission.RoomId = int64(rid)
	member.Permission.IsManage = set

	if err = m.db.SetRoomPermission(*member); err != nil {
		return err
	}

	return m.c.set(member)
}

// Kick
// 刪除會員所有房間快取
func (m *Member) Kick(uid string) ([]string, error) {
	keys, err := m.c.getKeys(uid)
	if err != nil {
		return nil, err
	}
	ok, err := m.c.delete(uid)
	if !ok {
		return nil, err
	}
	return keys, err
}

func (m *Member) GetKeys(uid string) ([]string, error) {
	return m.c.getKeys(uid)
}

func (m *Member) GetRoomKeys(uid string, rid int) ([]string, error) {
	return m.c.getRoomKeys(uid, rid)
}

func (m *Member) GetWs(uid string) (map[string]string, error) {
	return m.c.getWs(uid)
}

func (m *Member) GetMessageSession(uid string, rid int) (*models.Member, error) {
	member, err := m.GetByRoom(uid, rid)
	if err != nil {
		return nil, err
	}
	if member.Type == models.Guest {
		return nil, errors.ErrLogin
	}

	if member.Blockade() {
		return nil, errors.ErrBlockade
	}
	if member.Banned() {
		return nil, errors.ErrMemberNoMessage
	}

	ok, err := m.c.isBanned(uid, rid)
	if err != nil {
		return nil, err
	}
	if ok {
		return nil, errors.ErrMemberBanned
	}

	return member, nil
}

func (m *Member) GetSession(uid string) (*models.Member, error) {
	member, err := m.Get(uid)
	if err != nil {
		return nil, err
	}
	if member.Type == models.Guest {
		return nil, errors.ErrLogin
	}
	return member, nil
}

func (m *Member) Get(uid string) (*models.Member, error) {
	member, err := m.c.get(uid)
	if err != nil {
		if err == redis.Nil {
			return nil, errors.ErrLogin
		}
		return nil, err
	}
	return member, nil
}

// GetByRoom
// 取得會員在房間中資訊
func (m *Member) GetByRoom(uid string, rid int) (*models.Member, error) {
	u, err := m.c.getByRoom(uid, rid)

	if err == redis.Nil {

		u, err = m.db.Find(uid)

		if err == sql.ErrNoRows {
			return nil, errors.ErrNoMember
		}
		if err != nil {
			return nil, err
		}

	} else if err != nil {
		return nil, err
	}

	if u.Permission.RoomId != int64(rid) {
		u.Permission, _ = m.db.RoomPermission(u.Id, rid)
		u.Permission.RoomId = int64(rid)
		err = m.c.set(u)
	}

	return u, err
}

// GetStatus
// 取得會員狀態
func (m *Member) GetStatus(uid string, rid int) (*models.Member, error) {
	u, err := m.GetByRoom(uid, rid)
	if err != nil {
		return nil, err
	}

	if !u.Banned() {
		if u.Permission.IsBanned, err = m.c.isBanned(uid, rid); err != nil {
			return nil, err
		}
	}

	return u, nil
}

// GetUserName
// 取得會員名稱
func (m *Member) GetUserName(uid string) (string, error) {
	members, err := m.BatchGetUserNames([]string{uid})
	if err != nil {
		return "", err
	}
	return members[uid], nil
}

// BatchGetUserNames
// 批次取得會員帳號名稱(UserName)
func (m *Member) BatchGetUserNames(uid []string) (map[string]string, error) {
	name, err := m.c.getName(uid)
	if err != nil && err != redis.Nil {
		return nil, err
	}

	selectUid := make([]string, 0)
	for _, id := range uid {
		if _, ok := name[id]; !ok {
			selectUid = append(selectUid, id)
		}
	}

	if len(selectUid) == 0 {
		return name, nil
	}

	member, err := m.db.GetMembersByUid(selectUid)
	if err != nil {
		return nil, err
	}
	if name == nil {
		name = make(map[string]string, len(member))
	}

	cacheName := make(map[string]string, len(member))
	for _, v := range member {
		cacheName[v.Uid] = v.Name
		name[v.Uid] = v.Name
	}
	if err := m.c.setName(cacheName); err != nil {
		log.Error("set name cache for GetUserName", zap.Error(err), zap.Any("name", name))
	}
	return name, nil
}

// BatchGetMembers
// 批次取得會員資料 by id
func (m *Member) BatchGetMembers(id []int64) ([]models.Member, error) {
	return m.db.GetMembers(id)
}

// BatchGetMembersByUid
// 批次取得會員資料 by UID
func (m *Member) BatchGetMembersByUid(uid []string) ([]models.Member, error) {
	return m.db.GetMembersByUid(uid)
}

func (m *Member) Fetch(uid string) (*models.Member, error) {
	member, err := m.c.get(uid)
	if err == nil {
		return member, nil
	}
	if err != redis.Nil {
		return nil, err
	}

	dbMember, err := m.db.Find(uid)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrNoMember
		}
		return nil, err
	}
	if err := m.c.set(dbMember); err != nil {
		return nil, err
	}
	return member, nil
}

func (m *Member) Heartbeat(uid string) error {
	return m.c.refreshExpire(uid)
}

func (m *Member) Update(uid, name string, gender int32) error {
	u, err := m.db.Find(uid)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.ErrNoMember
		}
		return err
	}

	if u.Name != name || u.Gender != gender {
		u.Name = name
		u.Gender = gender
		if _, err := m.db.UpdateUser(u); err != nil {
			return err
		}
	}
	return nil
}

// SetLevel
// 更新會員等級
func (m *Member) SetLevel(uid string, rid int, level int) error {
	// Update level form DB
	u, err := m.db.Find(uid)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.ErrNoMember
		}
		return err
	}
	if u.Lv != level {
		u.Lv = level
		if _, err := m.db.UpdateUser(u); err != nil {
			return err
		}
	}

	// Update level form cache
	member, err := m.GetByRoom(uid, rid)
	if err != nil {
		return err
	}

	member.Lv = level

	return m.c.set(member)
}

type RedEnvelope struct {
	RoomId int

	// 單包金額 or 總金額 看Type種類決定·
	Amount int

	// 數量
	Count int

	// 紅包說明
	Message string

	// 紅包種類 拼手氣 or 普通
	Type string
}

func (m *Member) GiveRedEnvelope(uid, token string, redEnvelope RedEnvelope) (*models.Member, client.RedEnvelopeReply, error) {
	user, err := m.GetSession(uid)
	if err != nil {
		return nil, client.RedEnvelopeReply{}, err
	}

	log.Info("give red_envelope api", zap.String("uid", user.Uid), zap.Any("data", redEnvelope))

	give := client.RedEnvelope{
		RoomId:    redEnvelope.RoomId,
		Message:   redEnvelope.Message,
		Type:      redEnvelope.Type,
		Amount:    redEnvelope.Amount,
		Count:     redEnvelope.Count,
		ExpireMin: 120,
	}

	reply, err := m.cli.GiveRedEnvelope(give, token)
	if err != nil {
		return nil, client.RedEnvelopeReply{}, err
	}
	return user, reply, nil
}

type TakeResult struct {
	Name string `json:"name"`

	client.TakeEnvelopeReply
}

func (m *Member) TakeRedEnvelope(uid, token, redEnvelopeToken string) (TakeResult, error) {
	_, err := m.GetSession(uid)
	if err != nil {
		return TakeResult{}, err
	}
	reply, err := m.cli.TakeRedEnvelope(redEnvelopeToken, token)
	if err != nil {
		return TakeResult{}, err
	}

	var name string

	if reply.IsAdmin {
		name = RootName
	} else if reply.Uid != "" {
		if name, err = m.GetUserName(reply.Uid); err != nil {
			return TakeResult{}, err
		}
	}

	switch reply.Status {
	case client.TakeEnvelopeSuccess:
		reply.StatusMessage = "获得红包"
	case client.TakeEnvelopeReceived:
		reply.StatusMessage = "已经抢过了"
	case client.TakeEnvelopeGone:
		reply.StatusMessage = "手慢了，红包派完了"
	case client.TakeEnvelopeExpired:
		reply.StatusMessage = "红包已过期，不能抢"
	default:
		reply.StatusMessage = "不存在的红包"
	}
	return TakeResult{
		Name:              name,
		TakeEnvelopeReply: reply,
	}, nil
}

type RedEnvelopeDetail struct {
	client.RedEnvelopeInfo

	// 發紅包的會員名稱
	Name string `json:"name"`

	Members []MemberDetail `json:"members"`
}

type MemberDetail struct {
	client.MemberDetail

	// 搶走紅包會員的姓名
	Name string `json:"name"`
}

func (m *Member) GetRedEnvelopeDetail(orderId, authToken string) (RedEnvelopeDetail, error) {
	reply, err := m.cli.GetRedEnvelopeDetail(orderId, authToken)
	if err != nil {
		return RedEnvelopeDetail{}, err
	}

	var names map[string]string
	uids := make([]string, 0, len(reply.Members)+1)

	for _, v := range reply.Members {
		uids = append(uids, v.Uid)
	}
	if !reply.IsAdmin {
		uids = append(uids, reply.Uid)
	}

	members := []MemberDetail{}

	if len(uids) > 0 {
		members = make([]MemberDetail, 0, len(reply.Members)+1)
		if names, err = m.BatchGetUserNames(uids); err != nil {
			return RedEnvelopeDetail{}, err
		}
		for _, v := range reply.Members {
			members = append(members, MemberDetail{
				MemberDetail: v,
				Name:         names[v.Uid],
			})
		}
	}

	var name string

	if reply.IsAdmin {
		name = RootName
	} else {
		name = names[reply.Uid]
	}

	return RedEnvelopeDetail{
		RedEnvelopeInfo: reply.RedEnvelopeInfo,
		Name:            name,
		Members:         members,
	}, nil
}
