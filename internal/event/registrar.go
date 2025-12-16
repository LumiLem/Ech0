package event

// EventHandlers 事件处理器集合
type EventHandlers struct {
	wbd *WebhookDispatcher  // webhook 事件处理器
	dlr *DeadLetterResolver // 死信处理器
	fa  *FediverseAgent     // 联邦事件处理器
	bs  *BackupScheduler    // 备份事件调度器
	ap  *AgentProcessor     // Agent事件处理器
}

// NewEventHandlers 创建一个新的事件处理器集合
func NewEventHandlers(
	wbd *WebhookDispatcher,
	dlr *DeadLetterResolver,
	fa *FediverseAgent,
	bs *BackupScheduler,
	ap *AgentProcessor,
) *EventHandlers {
	return &EventHandlers{wbd: wbd, dlr: dlr, fa: fa, bs: bs, ap: ap}
}

// EventRegistrar 事件注册器
type EventRegistrar struct {
	eb IEventBus      // 事件总线
	eh *EventHandlers // 事件处理器集合
}

// NewEventRegistry 创建一个新的事件注册表
func NewEventRegistry(ebp func() IEventBus, eh *EventHandlers) *EventRegistrar {
	return &EventRegistrar{eb: ebp(), eh: eh}
}

// Register 注册事件处理函数
func (er *EventRegistrar) Register() error {
	var err error
	// 订阅死信事件
	err = er.eb.Subscribe(er.eh.dlr.Handle, EventTypeDeadLetterRetried) // 订阅死信事件，交给 DeadLetterResolver 处理
	if err != nil {
		return err
	}
	err = er.eb.Subscribe(er.eh.fa.Handle, EventTypeEchoCreated) // 订阅 EchoCreated 事件，交给 FediverseAgent 处理
	if err != nil {
		return err
	}
	err = er.eb.Subscribe(er.eh.bs.Handle, EventTypeUpdateBackupSchedule) // 订阅 UpdateBackupSchedule 事件，交给 BackupScheduler 处理
	if err != nil {
		return err
	}

	// 订阅事件组
	err = er.eb.Subscribes(
		er.eh.ap.Handle,
		EventTypeEchoCreated,
		EventTypeUserDeleted,
		EventTypeEchoUpdated,
	) // 订阅 Echo 事件组，交给 AgentProcessor 处理
	if err != nil {
		return err
	}
	// 订阅所有事件，交给 WebhookDispatcher 处理
	err = er.eb.SubscribeAll(er.eh.wbd.Handle, EventTypeDeadLetterRetried) // 订阅所有事件，交给 WebhookDispatcher 处理,但是排除死信事件
	if err != nil {
		return err
	}

	return err
}

// Wait 等待所有事件处理完成
func (er *EventRegistrar) Wait() {
	er.eh.wbd.Wait()
	er.eh.fa.Wait()
}
