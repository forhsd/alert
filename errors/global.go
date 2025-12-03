package errors

// // GlobalErrorCatcher 全局错误捕获器
// type GlobalErrorCatcher struct {
// 	alert   *storage.AlertLibrary
// 	timeout time.Duration
// }

// // NewGlobalErrorCatcher 创建全局错误捕获器
// func NewGlobalErrorCatcher(alert *storage.AlertLibrary) *GlobalErrorCatcher {
// 	return &GlobalErrorCatcher{
// 		alert:   alert,
// 		timeout: 5 * time.Second,
// 	}
// }

// // CatchAsync 捕获异步函数错误
// func (c *GlobalErrorCatcher) CatchAsync(fn func() error) {
// 	go func() {
// 		if err := fn(); err != nil {
// 			c.alert.ReportError(err)
// 		}
// 	}()
// }

// // WithContext 带上下文的错误捕获
// func (c *GlobalErrorCatcher) WithContext(ctx context.Context, fn func(context.Context) error) error {
// 	errChan := make(chan error, 1)

// 	go func() {
// 		defer func() {
// 			if r := recover(); r != nil {
// 				errChan <- fmt.Errorf("panic: %v", r)
// 			}
// 		}()
// 		errChan <- fn(ctx)
// 	}()

// 	select {
// 	case <-ctx.Done():
// 		return ctx.Err()
// 	case err := <-errChan:
// 		if err != nil {
// 			c.alert.ReportError(err)
// 		}
// 		return err
// 	case <-time.After(c.timeout):
// 		c.alert.ReportError(fmt.Errorf("function timeout after %v", c.timeout))
// 		return context.DeadlineExceeded
// 	}
// }
