package session

import (
	"log/slog"
	"reflect"
)

// Hook 常量，定义所有支持的钩子方法名。
const (
	BeforeQuery  = "BeforeQuery"
	AfterQuery   = "AfterQuery"
	BeforeInsert = "BeforeInsert"
	AfterInsert  = "AfterInsert"
	BeforeUpdate = "BeforeUpdate"
	AfterUpdate  = "AfterUpdate"
	BeforeDelete = "BeforeDelete"
	AfterDelete  = "AfterDelete"
)

// IBeforeQuery 查询前钩子。可用于记录日志、修改查询条件等。
type IBeforeQuery interface {
	BeforeQuery(s *Session) error
}

// IAfterQuery 查询后钩子。可用于对查询结果做后处理。
type IAfterQuery interface {
	AfterQuery(s *Session) error
}

// IBeforeInsert 插入前钩子。可用于自动填充字段（如创建时间）、数据校验等。
type IBeforeInsert interface {
	BeforeInsert(s *Session) error
}

// IAfterInsert 插入后钩子。可用于清理缓存、发送通知等。
type IAfterInsert interface {
	AfterInsert(s *Session) error
}

// IBeforeUpdate 更新前钩子。可用于自动填充更新时间等。
type IBeforeUpdate interface {
	BeforeUpdate(s *Session) error
}

// IAfterUpdate 更新后钩子。
type IAfterUpdate interface {
	AfterUpdate(s *Session) error
}

// IBeforeDelete 删除前钩子。可用于软删除拦截等。
type IBeforeDelete interface {
	BeforeDelete(s *Session) error
}

// IAfterDelete 删除后钩子。
type IAfterDelete interface {
	AfterDelete(s *Session) error
}

// CallMethod 调用模型上实现的钩子方法。
// 如果模型未实现对应的钩子接口，则静默跳过；如果钩子返回 error，则中断并返回该错误。
func (s *Session) CallMethod(method string, value any) error {
	// 确保拿到指针，这样才能匹配到指针接收者的接口实现
	p := reflect.ValueOf(value)
	if p.Kind() != reflect.Pointer {
		// 如果传入的是值，创建一个指针包装
		ptr := reflect.New(p.Type())
		ptr.Elem().Set(p)
		p = ptr
	}
	v := p.Interface()

	var err error
	switch method {
	case BeforeQuery:
		if h, ok := v.(IBeforeQuery); ok {
			err = h.BeforeQuery(s)
		}
	case AfterQuery:
		if h, ok := v.(IAfterQuery); ok {
			err = h.AfterQuery(s)
		}
	case BeforeInsert:
		if h, ok := v.(IBeforeInsert); ok {
			err = h.BeforeInsert(s)
		}
	case AfterInsert:
		if h, ok := v.(IAfterInsert); ok {
			err = h.AfterInsert(s)
		}
	case BeforeUpdate:
		if h, ok := v.(IBeforeUpdate); ok {
			err = h.BeforeUpdate(s)
		}
	case AfterUpdate:
		if h, ok := v.(IAfterUpdate); ok {
			err = h.AfterUpdate(s)
		}
	case BeforeDelete:
		if h, ok := v.(IBeforeDelete); ok {
			err = h.BeforeDelete(s)
		}
	case AfterDelete:
		if h, ok := v.(IAfterDelete); ok {
			err = h.AfterDelete(s)
		}
	}

	if err != nil {
		slog.Error("hook error", slog.String("method", method), slog.String("error", err.Error()))
	}
	return err
}
