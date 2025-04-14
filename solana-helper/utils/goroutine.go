package Utils

import (
	"sync"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"git.wkr.moe/web3/solana-helper/errcode"
)

func Parallel[T uint8 | uint16 | uint32 | uint64 | int8 | int16 | int32 | int64 | int | uint](ctx g.Ctx, times T, threads T, do func(i T, thread T) (err error)) {
	if threads == T(0) {
		threads = times
		if threads == T(0) {
			panic("并行线程数不应为 0")
		}
	}

	wg := sync.WaitGroup{}
	ch := make(chan T, threads)
	chIdle := make(chan T, threads)
	for thread := T(0); thread < threads; thread++ {
		chIdle <- thread
	}
	for i := T(0); i < times || times == T(0); i++ {
		thread := <-chIdle
		ch <- thread
		wg.Add(1)
		go func() {
			defer func() {
				wg.Done()
				<-ch
				chIdle <- thread
			}()
			err := do(i, thread)
			if err != nil {
				if !gerror.HasCode(err, errcode.IgnoreError) {
					g.Log().Errorf(ctx, "%+v", err)
				}
			}
		}()
	}
	wg.Wait()
}
