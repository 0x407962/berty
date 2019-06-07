import { observer } from 'mobx-react'
import { observe } from 'mobx'
import { Stream, StreamPagination } from './stream'
import { Unary } from './unary'
import { withStoreContext } from '@berty/store/context'
import { Component } from 'react'

@withStoreContext
@observer
export class ConfigEntity extends Component {
  render () {
    const { context, id, children } = this.props
    const entity = context.entity.config.get(id)
    if (entity) {
      return children(entity)
    }
    return null
  }
}

@withStoreContext
@observer
export class ContactEntity extends Component {
  render () {
    const { context, id, children } = this.props
    const entity = context.entity.contact.get(id)
    if (entity) {
      return children(entity)
    }
    return null
  }
}

@withStoreContext
@observer
export class DeviceEntity extends Component {
  render () {
    const { context, id, children } = this.props
    const entity = context.entity.device.get(id)
    if (entity) {
      return children(entity)
    }
    return null
  }
}

@withStoreContext
@observer
export class ConversationEntity extends Component {
  render () {
    const { context, id, children } = this.props
    const entity = context.entity.conversation.get(id)
    if (entity) {
      return children(entity)
    }
    return null
  }
}

@withStoreContext
@observer
export class ConversationMemberEntity extends Component {
  render () {
    const { context, id, children } = this.props
    const entity = context.entity.conversationMember.get(id)
    if (entity) {
      return children(entity)
    }
    return null
  }
}

@withStoreContext
@observer
export class EventEntity extends Component {
  render () {
    const { context, id, children } = this.props
    const entity = context.entity.event.get(id)
    if (entity) {
      return children(entity)
    }
    return null
  }
}

@withStoreContext
@observer
export class DevicePushConfigEntity extends Component {
  render () {
    const { context, id, children } = this.props
    const entity = context.entity.devicePushConfig.get(id)
    if (entity) {
      return children(entity)
    }
    return null
  }
}

@withStoreContext
@observer
export class DevicePushIdentifierEntity extends Component {
  render () {
    const { context, id, children } = this.props
    const entity = context.entity.devicePushIdentifier.get(id)
    if (entity) {
      return children(entity)
    }
    return null
  }
}

@withStoreContext
@observer
export class SenderAliasEntity extends Component {
  render () {
    const { context, id, children } = this.props
    const entity = context.entity.senderAlias.get(id)
    if (entity) {
      return children(entity)
    }
    return null
  }
}

@withStoreContext
export class IDServiceNode extends Unary {
  get service () {
    return this.props.context.node.service.id
  }
}

@withStoreContext
export class CommitLogStreamServiceNode extends Stream {
  get service () {
    return this.props.context.node.service.commitLogStream
  }
}

@withStoreContext
export class EventStreamServiceNode extends Stream {
  get service () {
    return this.props.context.node.service.eventStream
  }
}

@withStoreContext
export class EventListServiceNode extends Stream {
  get service () {
    return this.props.context.node.service.eventList
  }
}

@withStoreContext
class EventListServiceNodePagination extends StreamPagination {
  constructor (props, context) {
    super(props, context)
    observe(this.props.context.entity.event, this.observe)
  }

  get service () {
    return this.props.context.node.service.eventList
  }
}
EventListServiceNode.Pagination = EventListServiceNodePagination

@withStoreContext
export class EventUnseenServiceNode extends Stream {
  get service () {
    return this.props.context.node.service.eventUnseen
  }
}

@withStoreContext
class EventUnseenServiceNodePagination extends StreamPagination {
  constructor (props, context) {
    super(props, context)
    observe(this.props.context.entity.event, this.observe)
  }

  get service () {
    return this.props.context.node.service.eventUnseen
  }
}
EventUnseenServiceNode.Pagination = EventUnseenServiceNodePagination

@withStoreContext
export class GetEventServiceNode extends Unary {
  get service () {
    return this.props.context.node.service.getEvent
  }
}

@withStoreContext
export class EventSeenServiceNode extends Unary {
  get service () {
    return this.props.context.node.service.eventSeen
  }
}

@withStoreContext
export class EventRetryServiceNode extends Unary {
  get service () {
    return this.props.context.node.service.eventRetry
  }
}

@withStoreContext
export class ConfigPublicServiceNode extends Unary {
  get service () {
    return this.props.context.node.service.configPublic
  }
}

@withStoreContext
export class ConfigUpdateServiceNode extends Unary {
  get service () {
    return this.props.context.node.service.configUpdate
  }
}

@withStoreContext
export class ContactRequestServiceNode extends Unary {
  get service () {
    return this.props.context.node.service.contactRequest
  }
}

@withStoreContext
export class ContactAcceptRequestServiceNode extends Unary {
  get service () {
    return this.props.context.node.service.contactAcceptRequest
  }
}

@withStoreContext
export class ContactRemoveServiceNode extends Unary {
  get service () {
    return this.props.context.node.service.contactRemove
  }
}

@withStoreContext
export class ContactUpdateServiceNode extends Unary {
  get service () {
    return this.props.context.node.service.contactUpdate
  }
}

@withStoreContext
export class ContactListServiceNode extends Stream {
  get service () {
    return this.props.context.node.service.contactList
  }
}

@withStoreContext
class ContactListServiceNodePagination extends StreamPagination {
  constructor (props, context) {
    super(props, context)
    observe(this.props.context.entity.contact, this.observe)
  }

  get service () {
    return this.props.context.node.service.contactList
  }
}
ContactListServiceNode.Pagination = ContactListServiceNodePagination

@withStoreContext
export class ContactServiceNode extends Unary {
  get service () {
    return this.props.context.node.service.contact
  }
}

@withStoreContext
export class ContactCheckPublicKeyServiceNode extends Unary {
  get service () {
    return this.props.context.node.service.contactCheckPublicKey
  }
}

@withStoreContext
export class ConversationCreateServiceNode extends Unary {
  get service () {
    return this.props.context.node.service.conversationCreate
  }
}

@withStoreContext
export class ConversationUpdateServiceNode extends Unary {
  get service () {
    return this.props.context.node.service.conversationUpdate
  }
}

@withStoreContext
export class ConversationListServiceNode extends Stream {
  get service () {
    return this.props.context.node.service.conversationList
  }
}

@withStoreContext
class ConversationListServiceNodePagination extends StreamPagination {
  constructor (props, context) {
    super(props, context)
    observe(this.props.context.entity.conversation, this.observe)
  }

  get service () {
    return this.props.context.node.service.conversationList
  }
}
ConversationListServiceNode.Pagination = ConversationListServiceNodePagination

@withStoreContext
export class ConversationInviteServiceNode extends Unary {
  get service () {
    return this.props.context.node.service.conversationInvite
  }
}

@withStoreContext
export class ConversationExcludeServiceNode extends Unary {
  get service () {
    return this.props.context.node.service.conversationExclude
  }
}

@withStoreContext
export class ConversationAddMessageServiceNode extends Unary {
  get service () {
    return this.props.context.node.service.conversationAddMessage
  }
}

@withStoreContext
export class ConversationServiceNode extends Unary {
  get service () {
    return this.props.context.node.service.conversation
  }
}

@withStoreContext
export class ConversationMemberServiceNode extends Unary {
  get service () {
    return this.props.context.node.service.conversationMember
  }
}

@withStoreContext
export class ConversationReadServiceNode extends Unary {
  get service () {
    return this.props.context.node.service.conversationRead
  }
}

@withStoreContext
export class ConversationRemoveServiceNode extends Unary {
  get service () {
    return this.props.context.node.service.conversationRemove
  }
}

@withStoreContext
export class ConversationLastEventServiceNode extends Unary {
  get service () {
    return this.props.context.node.service.conversationLastEvent
  }
}

@withStoreContext
export class DevicePushConfigListServiceNode extends Unary {
  get service () {
    return this.props.context.node.service.devicePushConfigList
  }
}

@withStoreContext
export class DevicePushConfigCreateServiceNode extends Unary {
  get service () {
    return this.props.context.node.service.devicePushConfigCreate
  }
}

@withStoreContext
export class DevicePushConfigNativeRegisterServiceNode extends Unary {
  get service () {
    return this.props.context.node.service.devicePushConfigNativeRegister
  }
}

@withStoreContext
export class DevicePushConfigNativeUnregisterServiceNode extends Unary {
  get service () {
    return this.props.context.node.service.devicePushConfigNativeUnregister
  }
}

@withStoreContext
export class DevicePushConfigRemoveServiceNode extends Unary {
  get service () {
    return this.props.context.node.service.devicePushConfigRemove
  }
}

@withStoreContext
export class DevicePushConfigUpdateServiceNode extends Unary {
  get service () {
    return this.props.context.node.service.devicePushConfigUpdate
  }
}

@withStoreContext
export class HandleEventServiceNode extends Unary {
  get service () {
    return this.props.context.node.service.handleEvent
  }
}

@withStoreContext
export class GenerateFakeDataServiceNode extends Unary {
  get service () {
    return this.props.context.node.service.generateFakeData
  }
}

@withStoreContext
export class RunIntegrationTestsServiceNode extends Unary {
  get service () {
    return this.props.context.node.service.runIntegrationTests
  }
}

@withStoreContext
export class DebugPingServiceNode extends Unary {
  get service () {
    return this.props.context.node.service.debugPing
  }
}

@withStoreContext
export class DebugRequeueEventServiceNode extends Unary {
  get service () {
    return this.props.context.node.service.debugRequeueEvent
  }
}

@withStoreContext
export class DebugRequeueAllServiceNode extends Unary {
  get service () {
    return this.props.context.node.service.debugRequeueAll
  }
}

@withStoreContext
export class DeviceInfosServiceNode extends Unary {
  get service () {
    return this.props.context.node.service.deviceInfos
  }
}

@withStoreContext
export class AppVersionServiceNode extends Unary {
  get service () {
    return this.props.context.node.service.appVersion
  }
}

@withStoreContext
export class PeersServiceNode extends Unary {
  get service () {
    return this.props.context.node.service.peers
  }
}

@withStoreContext
export class ProtocolsServiceNode extends Unary {
  get service () {
    return this.props.context.node.service.protocols
  }
}

@withStoreContext
export class LogStreamServiceNode extends Stream {
  get service () {
    return this.props.context.node.service.logStream
  }
}

@withStoreContext
export class LogfileListServiceNode extends Stream {
  get service () {
    return this.props.context.node.service.logfileList
  }
}

@withStoreContext
export class LogfileReadServiceNode extends Stream {
  get service () {
    return this.props.context.node.service.logfileRead
  }
}

@withStoreContext
export class TestLogBackgroundErrorServiceNode extends Unary {
  get service () {
    return this.props.context.node.service.testLogBackgroundError
  }
}

@withStoreContext
export class TestLogBackgroundWarnServiceNode extends Unary {
  get service () {
    return this.props.context.node.service.testLogBackgroundWarn
  }
}

@withStoreContext
export class TestLogBackgroundDebugServiceNode extends Unary {
  get service () {
    return this.props.context.node.service.testLogBackgroundDebug
  }
}

@withStoreContext
export class TestPanicServiceNode extends Unary {
  get service () {
    return this.props.context.node.service.testPanic
  }
}

@withStoreContext
export class TestErrorServiceNode extends Unary {
  get service () {
    return this.props.context.node.service.testError
  }
}

@withStoreContext
export class MonitorBandwidthServiceNode extends Stream {
  get service () {
    return this.props.context.node.service.monitorBandwidth
  }
}

@withStoreContext
export class MonitorPeersServiceNode extends Stream {
  get service () {
    return this.props.context.node.service.monitorPeers
  }
}

@withStoreContext
export class GetListenAddrsServiceNode extends Unary {
  get service () {
    return this.props.context.node.service.getListenAddrs
  }
}

@withStoreContext
export class GetListenInterfaceAddrsServiceNode extends Unary {
  get service () {
    return this.props.context.node.service.getListenInterfaceAddrs
  }
}

@withStoreContext
export class Libp2PPingServiceNode extends Unary {
  get service () {
    return this.props.context.node.service.libp2PPing
  }
}

@withStoreContext
export class ServiceNode extends Component {
  static CommitLogStream = CommitLogStreamServiceNode
  static EventStream = EventStreamServiceNode
  static EventList = EventListServiceNode
  static EventUnseen = EventUnseenServiceNode
  static ContactList = ContactListServiceNode
  static ConversationList = ConversationListServiceNode
  static LogStream = LogStreamServiceNode
  static LogfileList = LogfileListServiceNode
  static LogfileRead = LogfileReadServiceNode
  static MonitorBandwidth = MonitorBandwidthServiceNode
  static MonitorPeers = MonitorPeersServiceNode

  render () {
    return context => this.props.children(context.node)
  }
}

@withStoreContext
export class Store extends Component {
  static Entity = {
    Config: ConfigEntity,
    Contact: ContactEntity,
    Device: DeviceEntity,
    Conversation: ConversationEntity,
    ConversationMember: ConversationMemberEntity,
    Event: EventEntity,
    DevicePushConfig: DevicePushConfigEntity,
    DevicePushIdentifier: DevicePushIdentifierEntity,
    SenderAlias: SenderAliasEntity,
  }

  static Node = {
    Service: ServiceNode,
  }

  render () {
    const { context } = this.props
    return this.props.children(context)
  }
}

export default Store
