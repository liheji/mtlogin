package util

import (
	"fmt"
	"sync"
	"time"

	"github.com/bwmarrin/snowflake"
)

const snowflakeNodeID int64 = 102

var (
	logIDOnce sync.Once
	logIDNode *snowflake.Node
	logIDErr  error
)

func NewLogID() string {
	logIDOnce.Do(func() {
		snowflake.Epoch = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli()
		logIDNode, logIDErr = snowflake.NewNode(snowflakeNodeID)
	})
	if logIDErr != nil {
		return fmt.Sprintf("%x", time.Now().UnixNano())
	}
	return fmt.Sprintf("%x", logIDNode.Generate().Int64())
}
