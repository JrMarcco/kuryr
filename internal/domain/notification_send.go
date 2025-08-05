package domain

// SendResult 消息发送结果领域对象
type SendResult struct {
	NotificationId uint64
	SendStatus     SendStatus
}

// SendResp 消息请求响应领域对象
type SendResp struct {
	Result SendResult
}

// BatchSendResp 消息批量发送请求响应领域对象
type BatchSendResp struct {
	Results []SendResult
}

// BatchAsyncSendResp 批量异步发送请求响应领域对象
type BatchAsyncSendResp struct {
	NotificationIds []uint64
}
