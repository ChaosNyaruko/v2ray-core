logHandler: RegisterHandler()

RegisterHandler has a Handler interface, implemening Handle(),
and can be overwritten


Handler can be created by createHandler(), which uses "handlerCreatorMap" to find the "creator", and calls "creator(logType, options)" to return a Handler

handlerCreatorMap <- assigned in app/log/log_creator.go, RegisterHandlerCreator  
this map is assigned in log_creator.go, using log.NewLogger, because it is also a "Handler"


creator: func (LogType, HandlerCreatorOptions) (log.Handler, error)

