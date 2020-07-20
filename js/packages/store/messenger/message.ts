import { composeReducers } from 'redux-compose'
import { CaseReducer, PayloadAction, createSlice } from '@reduxjs/toolkit'
import { all, put, takeEvery, select, call } from 'redux-saga/effects'
import moment from 'moment'
import { berty } from '@berty-tech/api'
import { makeDefaultCommandsSagas, strToBuf, bufToStr, jsonToBuf, bufToJSON } from '../utils'

import * as protocol from '../protocol'
import * as conversation from './conversation'

import { UserMessage, GroupInvitation, AppMessageType, AppMessage, Acknowledge } from './AppMessage'

export type StoreUserMessage = UserMessage & { isMe: boolean; acknowledged: boolean }
export type StoreGroupInvitation = GroupInvitation & { isMe: boolean }

export type StoreMessage = StoreUserMessage | StoreGroupInvitation

export type Entity = {
	id: string
	receivedDate: number
} & StoreMessage

export type State = {
	events: Array<Event>
	aggregates: { [key: string]: Entity | undefined }
	ackBacklog: { [key: string]: true | undefined }
}

export type GlobalState = {
	messenger: {
		message: State
	}
}

export namespace Command {
	export type Delete = { id: string }
	export type Send = AppMessage & { id: string }
	export type Hide = void
	export type SendToAll = void
}

export namespace Query {
	export type List = {}
	export type Get = { id: string }
	export type GetLength = void
	export type GetList = { list: Entity['id'][] }
	export type Search = { searchText: string; list: Entity['id'][] }
	export type SearchOne = { searchText: string; id: Entity['id'] }
}

export namespace Event {
	export type Deleted = { aggregateId: string }
	export type Received = {
		aggregateId: string
		message: AppMessage
		receivedDate: number
		isMe: boolean
		memberPk?: string
	}
	export type Hidden = { aggregateId: string }
}

type SimpleCaseReducer<P> = CaseReducer<State, PayloadAction<P>>

export type CommandsReducer = {
	delete: SimpleCaseReducer<Command.Delete>
	send: SimpleCaseReducer<Command.Send>
	hide: SimpleCaseReducer<Command.Hide>
	sendToAll: SimpleCaseReducer<Command.SendToAll>
}

type FakeConfig = {
	body: string
}

const FAKE_MESSAGES_CONFIG: FakeConfig[] = [
	{ body: 'Welcome to Berty' },
	{ body: 'Hey, nice to see you here' },
	{ body: 'Get out of here! It has all the cryptos!' },
	{ body: 'Hey, I just arrived in New York' },
]

const FAKE_MESSAGES: Entity[] = FAKE_MESSAGES_CONFIG.map((fc, i) => {
	return {
		type: AppMessageType.UserMessage,
		id: `fake_${i}`,
		body: fc.body,
		attachments: [],
		sentDate: Date.now(),
		receivedDate: Date.now(),
		isMe: false,
		acknowledged: false,
	}
})

export type QueryReducer = {
	list: (state: GlobalState, query: Query.List) => Entity[]
	get: (state: GlobalState, query: Query.Get) => Entity | undefined
	getLength: (state: GlobalState) => number
	getList: (state: GlobalState, query: Query.GetList) => Entity[]
	search: (state: GlobalState, query: Query.Search) => StoreUserMessage[]
	searchOne: (state: GlobalState, query: Query.SearchOne) => Entity | undefined
}

export type EventsReducer = {
	deleted: SimpleCaseReducer<Event.Deleted>
	received: SimpleCaseReducer<Event.Received>
	hidden: SimpleCaseReducer<Event.Hidden>
}

export type Transactions = {
	[K in keyof CommandsReducer]: CommandsReducer[K] extends SimpleCaseReducer<infer TPayload>
		? (payload: TPayload) => Generator
		: never
}

const initialState: State = {
	events: [],
	aggregates: {},
	ackBacklog: {},
}

const commandHandler = createSlice<State, CommandsReducer>({
	name: 'messenger/message/command',
	initialState,
	reducers: {
		delete: (state) => state,
		send: (state) => state,
		hide: (state) => state,
		sendToAll: (state) => state,
	},
})

const eventHandler = createSlice<State, EventsReducer>({
	name: 'messenger/message/event',
	initialState,
	reducers: {
		received: (state, { payload: { aggregateId, message, receivedDate, isMe, memberPk } }) => {
			if (state.aggregates[aggregateId]) {
				return state
			}
			switch (message.type) {
				case AppMessageType.UserMessage:
					const hasAckInBacklog = !!state.ackBacklog[aggregateId]
					if (hasAckInBacklog) {
						delete state.ackBacklog[aggregateId]
					}
					const templateMsg = {
						id: aggregateId,
						type: message.type,
						body: message.body,
						isMe,
						attachments: [],
						sentDate: message.sentDate,
						acknowledged: hasAckInBacklog,
						receivedDate,
					}
					state.aggregates[aggregateId] = memberPk ? { ...templateMsg, memberPk } : templateMsg
					break
				case AppMessageType.UserReaction:
					// todo: append reaction to message
					break
				case AppMessageType.GroupInvitation:
					state.aggregates[aggregateId] = {
						id: aggregateId,
						type: message.type,
						group: message.group,
						isMe,
						receivedDate,
						name: message.name,
					}
					break
				case AppMessageType.Acknowledge:
					if (!isMe) {
						const target = state.aggregates[message.target]
						if (!target) {
							state.ackBacklog[message.target] = true
						} else if (target.type === AppMessageType.UserMessage && target.isMe) {
							target.acknowledged = true
						}
					}
					break
			}
			return state
		},
		hidden: (state) => state,
		deleted: (state, { payload }) => {
			delete state.aggregates[payload.aggregateId]
			return state
		},
	},
})

const getAggregatesWithFakes = (state: GlobalState) => {
	// TODO: optimize
	const result: { [key: string]: Entity | undefined } = { ...state.messenger.message.aggregates }
	for (const fake of FAKE_MESSAGES) {
		result[fake.id] = fake
	}
	return result
}

export const reducer = composeReducers(commandHandler.reducer, eventHandler.reducer)
export const commands = commandHandler.actions
export const events = eventHandler.actions
export const queries: QueryReducer = {
	list: (state) => Object.values(getAggregatesWithFakes(state)) as Entity[],
	get: (state, { id }) => getAggregatesWithFakes(state)[id],
	getLength: (state) => Object.keys(getAggregatesWithFakes(state)).length,
	getList: (state, { list }) => {
		if (!list) {
			return []
		}
		const messages = list.map((id) => {
			const ret = state.messenger.message.aggregates[id]
			return ret
		})
		return messages as Entity[]
	},
	search: (state, { searchText, list }) => {
		const messages = list
			.map((id) => state.messenger.message.aggregates[id] as StoreUserMessage)
			.filter((message) => message && message.type === AppMessageType.UserMessage)
			.filter(
				(message) =>
					message && message.body && message.body.toLowerCase().includes(searchText.toLowerCase()),
			)
		return !searchText ? [] : messages
	},
	searchOne: (state, { searchText, id }) => {
		const message: Entity | undefined = state.messenger.message.aggregates[id]
		return message &&
			message.type === AppMessageType.UserMessage &&
			message?.body &&
			message?.body.toLowerCase().includes(searchText.toLowerCase())
			? message
			: undefined
	},
}

export enum CommandsMessageType {
	Help = 'help',
	DebugGroup = 'debug-group',
	SendMessage = 'send-message',
}

const addCommandMessage = function* (conv: any, title: any, response: any) {
	const index: any = yield select((state) => queries.getLength(state))
	const aggregateId = getAggregateId({
		accountId: conv.id,
		index: index.toString(),
	})
	const message: UserMessage = {
		type: AppMessageType.UserMessage,
		body: '/' + title + '\n\n' + response,
		attachments: [],
		sentDate: Date.now(),
	}
	yield put(
		events.received({
			aggregateId,
			message,
			receivedDate: Date.now(),
			isMe: true,
		}),
	)
	yield call(conversation.transactions.addMessage, {
		aggregateId: conv.pk,
		messageId: aggregateId,
		isMe: true,
	})
}

export const getCommandsMessage = () => {
	const arr = {
		[CommandsMessageType.Help]: {
			help: 'Displays all commands',
			run: function* (context: any, args: any) {
				const response =
					'/help					Show this command\n' +
					'/debug-group				Indicate 1to1 connection\n' +
					'/send-message [message]	Send message\n'
				yield addCommandMessage(context.conv, CommandsMessageType.Help, response)
				return
			},
		},
		[CommandsMessageType.DebugGroup]: {
			help: 'List peers in this group',
			run: function* (context: any, args: any) {
				try {
					const group: any = yield* protocol.transactions.client.debugGroup({
						id: context.conv.accountId,
						groupPk: strToBuf(context.conv.pk),
					}) // does not support multi devices per account
					const response = group.peerIds.length
						? 'You are connected with this peer !'
						: 'You are not connected with this peer ...'
					yield addCommandMessage(context.conv, CommandsMessageType.DebugGroup, response)
					return
				} catch (error) {
					return error
				}
			},
		},
		[CommandsMessageType.SendMessage]: {
			help: 'Send a message',
			run: function* (context: any, args: any) {
				try {
					// TODO: implem minimist and put this const in args
					const body = context.payload.body.substr(CommandsMessageType.SendMessage.length + 2) // 2 = slash + space before the message
					const response = !body ? 'Invalid arguments ...' : 'You have sent a message !'
					yield addCommandMessage(context.conv, CommandsMessageType.SendMessage, response)
					if (body) {
						const userMessage: UserMessage = {
							type: AppMessageType.UserMessage,
							body,
							attachments: [],
							sentDate: Date.now(),
						}

						yield* protocol.transactions.client.appMessageSend({
							id: context.conv.accountId,
							groupPk: strToBuf(context.conv.pk), // need to set the pk in conv handlers
							payload: jsonToBuf(userMessage),
						})
					}
					return
				} catch (error) {
					return error
				}
			},
		},
	}
	return arr
}

export const isCommandMessage = (message: string) => {
	const commands: any = getCommandsMessage()
	// index for simple command
	let index = message.split('\n')[0]
	let cmd = commands[index.substr(1)]
	if (cmd && index.substr(0, 1) === '/') {
		return cmd
	}
	// index for command w/ args
	index = message.split(' ')[0]
	cmd = commands[index.substr(1)]
	if (cmd && index.substr(0, 1) === '/') {
		return cmd
	}
	return null
}

export const getAggregateId: (kwargs: { accountId: string; index: string }) => string = ({
	accountId,
	index,
}) => bufToStr(Buffer.concat([Buffer.from(accountId, 'utf-8'), Buffer.from(index, 'utf-8')]))

export const transactions: Transactions = {
	delete: function* ({ id }) {
		yield put(
			events.deleted({
				aggregateId: id,
			}),
		)
	},
	send: function* (payload) {
		// Recup the conv
		const conv = (yield select((state) => conversation.queries.get(state, { id: payload.id }))) as
			| conversation.Entity
			| undefined
		if (!conv) {
			return
		}

		if (payload.type === AppMessageType.UserMessage) {
			const cmd = isCommandMessage(payload.body)
			if (cmd) {
				const context = {
					conv,
					payload,
				}
				const args = {}
				yield cmd?.run(context, args)
			} else {
				const message: UserMessage = {
					type: AppMessageType.UserMessage,
					body: payload.body,
					attachments: payload.attachments,
					sentDate: Date.now(),
				}

				yield* protocol.transactions.client.appMessageSend({
					groupPk: strToBuf(conv.pk), // need to set the pk in conv handlers
					payload: jsonToBuf(message),
				})
			}
		}
	},
	sendToAll: function* () {
		// Recup the conv
		const conv = (yield select((state) => conversation.queries.list(state))) as
			| Array<conversation.Entity | conversation.FakeConversation>
			| undefined
		if (!conv || (conv && !conv.length)) {
			return
		}
		for (let i = 0; i < conv.length; i++) {
			if (conv[i].kind !== 'fake') {
				const message: UserMessage = {
					type: AppMessageType.UserMessage,
					body: `Test, ${moment().format('MMMM Do YYYY, h:mm:ss a')}`,
					attachments: [],
					sentDate: Date.now(),
				}
				yield* protocol.transactions.client.appMessageSend({
					groupPk: strToBuf(conv[i].pk), // need to set the pk in conv handlers
					payload: jsonToBuf(message),
				})
			}
		}
	},
	hide: function* () {
		// TODO: hide a message
	},
}

export function* orchestrator() {
	yield all([
		...makeDefaultCommandsSagas(commands, transactions),
		takeEvery('protocol/GroupMessageEvent', function* (
			action: PayloadAction<berty.types.v1.IGroupMessageEvent & { aggregateId: string }>,
		) {
			// create an id for the message
			const idBuf = action.payload.eventContext?.id
			if (!idBuf) {
				return
			}
			const groupPkBuf = action.payload.eventContext?.groupPk
			if (!groupPkBuf) {
				return
			}
			if (!action.payload.message) {
				return
			}
			const message: AppMessage = bufToJSON(action.payload.message) // <--- Not secure
			const aggregateId = bufToStr(idBuf)
			// create the message entity
			const existingMessage = (yield select((state) => queries.get(state, { id: aggregateId }))) as
				| Entity
				| undefined
			if (existingMessage) {
				return
			}
			// Reconstitute the convId
			const convId = bufToStr(groupPkBuf)
			// Recup the conv
			const conv = (yield select((state) => conversation.queries.get(state, { id: convId }))) as
				| conversation.Entity
				| undefined
			if (!conv) {
				return
			}

			const msgDevicePk = action.payload.headers?.devicePk

			let memberPk: string | undefined
			if (msgDevicePk) {
				const msgDevicePkStr = bufToStr(msgDevicePk)
				const groups = yield select((state) => state.groups)
				const { membersDevices } = groups[conv.pk] || { membersDevices: {} }
				const [pk] =
					Object.entries((membersDevices as { [key: string]: any }) || {}).find(
						([, devicePks]) => devicePks && devicePks.some((pk: any) => pk === msgDevicePkStr),
					) || []
				memberPk = pk
			}

			const groupInfo: berty.types.v1.GroupInfo.IReply = yield call(
				protocol.transactions.client.groupInfo,
				{
					id: action.payload.aggregateId,
					groupPk: groupPkBuf,
					contactPk: new Uint8Array(),
				},
			)
			let isMe = false
			if (msgDevicePk && groupInfo?.devicePk) {
				// TODO: multiple devices support
				isMe = Buffer.from(msgDevicePk).equals(Buffer.from(groupInfo.devicePk))
			}

			yield put(
				events.received({
					aggregateId,
					message,
					receivedDate: Date.now(),
					isMe,
					memberPk,
				}),
			)

			// Add received message in store

			if (
				message.type === AppMessageType.UserMessage ||
				message.type === AppMessageType.GroupInvitation
			) {
				yield call(conversation.transactions.addMessage, {
					aggregateId: bufToStr(groupPkBuf),
					messageId: aggregateId,
					isMe,
				})
			}

			if (message.type === AppMessageType.UserMessage) {
				// add message to corresponding conversation

				if (!isMe) {
					// send acknowledgment
					const acknowledge: Acknowledge = {
						type: AppMessageType.Acknowledge,
						target: aggregateId,
					}
					yield call(protocol.transactions.client.appMessageSend, {
						groupPk: groupPkBuf,
						payload: jsonToBuf(acknowledge),
					})
				}
			}
		}),
	])
}
