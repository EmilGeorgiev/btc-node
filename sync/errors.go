package sync

import "errors"

var (
	ErrNotFound                    = errors.New("not found")
	ErrFailedToGetLastBlock        = errors.New("failed to get the last block from DB ")
	ErrFailedToCreateMsgGetHeaders = errors.New("failed to create MSG GetHeaders ")
	ErrFailedToSendMsgGetHeaders   = errors.New("failed to send MS GetHeaders ")
)
