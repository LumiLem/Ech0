package async

import (
	"context"
	"sync"

	logUtil "github.com/lin-snow/ech0/internal/util/log"
	"go.uber.org/zap"
)

// WorkerPool 是一个可复用的通用异步任务池
type WorkerPool struct {
	workerCount int                // 并发数
	jobs        chan func() error  // 任务通道
	wg          sync.WaitGroup     // 用于等待所有任务完成
	ctx         context.Context    // 上下文
	cancel      context.CancelFunc // 取消函数
}

// NewWorkerPool 创建一个新的 WorkerPool
func NewWorkerPool(workerCount, jobQueueSize int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	workerPool := &WorkerPool{
		workerCount: workerCount,
		jobs:        make(chan func() error, jobQueueSize),
		ctx:         ctx,
		cancel:      cancel,
	}
	workerPool.start()
	return workerPool
}

// Start 启动工作池
func (p *WorkerPool) start() {
	for i := 0; i < p.workerCount; i++ {
		go func() {
			for {
				select {
				case job, ok := <-p.jobs:
					if !ok {
						return
					}
					p.wg.Add(1)
					if err := job(); err != nil {
						logUtil.GetLogger().
							Error("worker job failed", zap.String("err", err.Error()))
					}
					p.wg.Done()
				case <-p.ctx.Done():
					return
				}
			}
		}()
	}
}

// Submit 提交一个任务到工作池
func (p *WorkerPool) Submit(job func() error) {
	select {
	case p.jobs <- job:
		// 任务已提交
	case <-p.ctx.Done():
		// 如果上下文已取消，则不提交任务
	}
}

// Wait 等待所有任务完成
func (p *WorkerPool) Wait() {
	p.wg.Wait()
}

// Stop 停止工作池
func (p *WorkerPool) Stop() {
	p.cancel()
	close(p.jobs)
	p.Wait() // 确保所有任务完成
}
