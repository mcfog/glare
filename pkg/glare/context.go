package glare

import "github.com/secmask/go-redisproto"

type requestContext struct {
	writer  *redisproto.Writer
	command *redisproto.Command
	offset  int
}

func (ctx *requestContext) Get(idx int) []byte {
	return ctx.command.Get(idx + ctx.offset)
}

func (ctx *requestContext) Shift() []byte {
	if ctx.CountRemainArg() <= 0 {
		return nil
	}

	offset := ctx.offset
	ctx.offset += 1
	return ctx.command.Get(offset)
}

func (ctx *requestContext) MatchAndShift(cmd string) bool {
	if ctx.CountRemainArg() <= 0 {
		return false
	}

	if string(ctx.Get(0)) == cmd {
		ctx.Shift()
		return true
	}

	return false
}

func (ctx *requestContext) GetWriter() *redisproto.Writer {
	return ctx.writer
}

func (ctx *requestContext) CountRemainArg() int {
	return ctx.command.ArgCount() - ctx.offset
}
