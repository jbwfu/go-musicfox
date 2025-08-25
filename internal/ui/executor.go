package ui

import (
	"github.com/anhoder/foxful-cli/model"
	_struct "github.com/go-musicfox/go-musicfox/utils/struct"
)

// CoreFunc 是操作的核心逻辑函数，它接收 Netease 实例作为参数。
type CoreFunc func(m *Netease) model.Page

// Operation 代表一个可执行的操作单元。
type Operation struct {
	m           *Netease
	coreFunc    CoreFunc
	needsAuth   bool
	showLoading bool
}

// NewOperation 创建一个新操作。
// 参数 m 是 Netease 主模型，coreFunc 是要执行的核心业务逻辑。
func NewOperation(m *Netease, coreFunc CoreFunc) *Operation {
	return &Operation{
		m:        m,
		coreFunc: coreFunc,
	}
}

// SetNeedsAuth 声明该操作需要用户登录。
// 返回 Operation 指针以支持链式调用。
func (op *Operation) SetNeedsAuth(needs bool) *Operation {
	op.needsAuth = needs
	return op
}

// SetShowLoading 声明该操作在执行期间应显示加载动画。
// 返回 Operation 指针以支持链式调用。
func (op *Operation) SetShowLoading(show bool) *Operation {
	op.showLoading = show
	return op
}

// Execute 按照配置执行操作。
// 它会按顺序处理：加载动画 -> 认证检查 -> 核心逻辑。
func (op *Operation) Execute() model.Page {
	// 1. Loading "中间件"
	if op.showLoading {
		// 使用 defer 确保加载动画在操作完成（无论成功与否）后关闭
		loading := model.NewLoading(op.m.MustMain())
		loading.Start()
		defer loading.Complete()
	}

	// 2. Auth "中间件"
	if op.needsAuth {
		if _struct.CheckUserInfo(op.m.user) == _struct.NeedLogin {
			// 如果需要登录但用户未登录，则跳转到登录页面。
			// 登录成功后的回调函数是再次执行当前操作，以确保流程的延续。
			page, _ := op.m.ToLoginPage(func() model.Page {
				// 注意：这里我们重新执行 op.Execute()
				// 这可以确保在登录成功后，loading 和其他检查可以再次正确执行。
				return op.Execute()
			})
			return page
		}
	}

	// 3. 一切就绪，执行真正的核心逻辑
	return op.coreFunc(op.m)
}
