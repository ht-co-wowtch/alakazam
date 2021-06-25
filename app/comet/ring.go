package comet

import (
	"gitlab.com/jetfueltw/cpw/alakazam/app/comet/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/app/comet/pb"
)

//TODO: refactor 使用官方環形api

// 用於控制讀寫異步grpc.Proto的環型Pool
type Ring struct {
	// 讀的游標
	rp uint64

	// data長度
	num uint64

	// 用於切換異步grpc.Proto游標
	mask uint64

	// 寫的游標
	wp uint64

	// 多個grpc Proto結構
	data []pb.Proto
}

func (r *Ring) init(num uint64) {
	// 如果num非2的公倍數則轉成2的公倍數
	// 因為需保證讀寫游標與mask的&操作是可以對Proto做循環取得
	if num&(num-1) != 0 {

		for num&(num-1) != 0 {
			num &= (num - 1)
		}
		num = num << 1
	}
	r.data = make([]pb.Proto, num)
	r.num = num
	r.mask = r.num - 1
}

// 取用於寫的grpc.Proto，如果讀跟寫游標相等代表沒有可以讀的Proto
func (r *Ring) Get() (*pb.Proto, error) {
	if r.rp == r.wp {
		return nil, errors.ErrRingEmpty
	}
	return &r.data[r.rp&r.mask], nil
}

// 讀游標++
func (r *Ring) GetAdv() {
	r.rp++
}

func (r *Ring) Init(num int) {
	r.init(uint64(num))
}

// 取用於寫的grpc.Proto，讀跟寫的游標不可相差大於等於Proto數量
// 需要要防寫的速度大於讀的速度時會把已寫未讀的Proto做覆蓋
// 假設Proto數量是6個(A,B,C,D,E,F)，w(寫游標)，r(讀游標)
// ====================================================
// 沒有可讀Proto
//
//		r
//		↓
// 		A	B	C	D	E	F
//		↑
//		w
// ====================================================
// 可讀Proto
//
//		r
//		↓
// 		A	B	C	D	E	F
//			↑
//			w
// ====================================================
// 不可寫Proto
//
//		r
//		↓
// 		A	B	C	D	E	F
//							↑
//							w
//
func (r *Ring) Set() (*pb.Proto, error) {
	if r.wp-r.rp >= r.num {
		return nil, errors.ErrRingFull
	}
	return &r.data[r.wp&r.mask], nil
}

// 寫游標++
func (r *Ring) SetAdv() {
	r.wp++
}

// 重置讀寫游標
func (r *Ring) Reset() {
	//need to check
	r.wp = 0
	r.rp = 0
}
