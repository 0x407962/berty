import { composeReducers } from 'redux-compose'
import { CaseReducer, PayloadAction, createSlice } from '@reduxjs/toolkit'
import { all, takeLeading, put, takeEvery, select } from 'redux-saga/effects'
import { Buffer } from 'buffer'

import * as protocol from '../protocol'
import { conversation } from '../chat'
import {
	UserMessage,
	UserReaction,
	GroupInvitation,
	AppMessageType,
	AppMessage,
} from './AppMessage'

export type Entity = {
	id: string
} & (UserMessage | UserReaction | GroupInvitation)

export type Event = {
	id: string
	version: number
	aggregateId: string
}

export type State = {
	events: Array<Event>
	aggregates: { [key: string]: Entity }
}

export type GlobalState = {
	chat: {
		message: State
	}
}

export namespace Command {
	export type Delete = { id: string }
	export type Send = AppMessage & { id: string }
	export type Hide = void
}

export namespace Query {
	export type List = {}
	export type Get = { id: string }
	export type GetLength = void
}

export namespace Event {
	export type Deleted = { aggregateId: string }
	export type Sent = { aggregateId: string; payload: UserMessage | UserReaction | GroupInvitation }
	export type Hidden = { aggregateId: string }
}

type SimpleCaseReducer<P> = CaseReducer<State, PayloadAction<P>>

export type CommandsReducer = {
	delete: SimpleCaseReducer<Command.Delete>
	send: SimpleCaseReducer<Command.Send>
	hide: SimpleCaseReducer<Command.Hide>
}

export type QueryReducer = {
	list: (state: GlobalState, query: Query.List) => Array<Entity>
	get: (state: GlobalState, query: Query.Get) => Entity
	getLength: (state: GlobalState) => number
}

export type EventsReducer = {
	deleted: SimpleCaseReducer<Event.Deleted>
	sent: SimpleCaseReducer<Event.Sent>
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
}

const commandHandler = createSlice<State, CommandsReducer>({
	name: 'chat/message/command',
	initialState,
	reducers: {
		delete: (state) => state,
		send: (state) => state,
		hide: (state) => state,
	},
})

const eventHandler = createSlice<State, EventsReducer>({
	name: 'chat/message/event',
	initialState,
	reducers: {
		sent: (state, { payload }) => {
			if (payload.payload.type === AppMessageType.UserMessage) {
				console.log('received message', payload)
				state.aggregates[payload.aggregateId] = {
					id: payload.aggregateId,
					type: payload.payload.type,
					body: payload.payload.body,
					attachments: [],
				}
			} else if (payload.payload.type === AppMessageType.UserReaction) {
				state.aggregates[payload.aggregateId] = {
					id: payload.aggregateId,
					type: payload.payload.type,
					emoji: payload.payload.emoji,
				}
			} else if (payload.payload.type === AppMessageType.GroupInvitation) {
				state.aggregates[payload.aggregateId] = {
					id: payload.aggregateId,
					type: payload.payload.type,
					groupPk: payload.payload.groupPk,
				}
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

export const reducer = composeReducers(commandHandler.reducer, eventHandler.reducer)
export const commands = commandHandler.actions
export const events = eventHandler.actions
export const queries: QueryReducer = {
	list: (state) => Object.values(state.chat.message.aggregates),
	get: (state, { id }) => state.chat.message.aggregates[id],
	getLength: (state) => Object.keys(state.chat.message.aggregates).length,
}

const getAggregateId: (kwargs: { accountId: string; groupPk: Uint8Array }) => string = ({
	accountId,
	groupPk,
}) => Buffer.concat([Buffer.from(accountId, 'utf-8'), Buffer.from(groupPk)]).toString('base64')

export const transactions: Transactions = {
	delete: function*({ id }) {
		yield put(
			events.deleted({
				aggregateId: id,
			}),
		)
	},
	send: function*(payload) {
		// Recup the conv
		const conv = (yield select((state) =>
			conversation.queries.get(state, { id: payload.id }),
		)) as conversation.Entity
		if (!conv) {
			return
		}

		if (payload.type === AppMessageType.UserMessage) {
			const message: UserMessage = {
				type: AppMessageType.UserMessage,
				body: payload.body,
				attachments: payload.attachments,
			}

			yield* protocol.transactions.client.appMessageSend({
				id: conv.accountId,
				groupPk: Buffer.from(conv.pk, 'utf-8'),
				payload: Buffer.from(JSON.stringify(message), 'utf-8'),
			})
		}
	},
	hide: function*() {
		// TODO: hide a message
	},
}

export function* orchestrator() {
	yield all([
		takeLeading(commands.delete, function*({ payload }) {
			yield* transactions.delete(payload)
		}),
		takeLeading(commands.send, function*({ payload }) {
			yield* transactions.send(payload)
		}),
		takeLeading(commands.hide, function*({ payload }) {
			yield* transactions.hide(payload)
		}),
		takeEvery('protocol/GroupMessageEvent', function*(action) {
			const message = JSON.parse(new Buffer(action.payload.message).toString('utf-8'))
			// create an id for the message
			const aggregateId = Buffer.from(action.payload.eventContext.id).toString('utf-8')
			// create the message entity
			const existingMessage = (yield select((state) => queries.get(state, { id: aggregateId }))) as
				| Entity
				| undefined
			if (existingMessage) {
				return
			}
			yield put(
				events.sent({
					aggregateId,
					payload: message,
				}),
			)
			// add message to correspondant conversation
			yield* conversation.transactions.addMessage({
				aggregateId: getAggregateId({
					accountId: action.payload.aggregateId,
					groupPk: action.payload.eventContext.groupPk,
				}),
				messageId: aggregateId,
			})
		}),
	])
}
