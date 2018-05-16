package glare

func handleGrpc(ctx *requestContext, svr *Server) (err error) {
	switch {
	case ctx.MatchAndShift("REQUEST"):
		argc := ctx.CountRemainArg() - 1
		if argc < 0 {
			err = ctx.GetWriter().WriteError("Payload not found")
			return
		}
		argv := make([][]byte, argc)
		payload := ctx.Get(argc)
		for i := range argv {
			argv[i] = ctx.Get(i)
		}

		response, handleErr := svr.handlerRequest(argv, payload)
		if handleErr != nil {
			err = ctx.GetWriter().WriteError(handleErr.Error())
		} else {
			err = ctx.GetWriter().WriteBulk(response)
		}

		return
	default:
		err = ctx.GetWriter().WriteError("Command not support")
		return
	}
}
