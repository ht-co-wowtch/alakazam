package logic

func (l *Logic) SetBlockade(uid, remark string) error {
	// TODO 待實作 從redis hash table 找出status並改成封鎖狀態
	// business.Blockade
	return nil
}

func (l *Logic) RemoveBlockade(uid string) error {
	// TODO 待實作 從redis hash table 找出status並寫回原來status
	// 原來的status要從DB拿，先暫時預設用business.PlayDefaultPermission
	return nil
}
