import { useContext, useMemo } from 'react'
import { MsgrContext, useMsgrContext } from './context'
import flatten from 'lodash/flatten'
import { messenger as messengerpb } from '@berty-tech/api/index.js'

const AppMessageType = messengerpb.AppMessage.Type

const listHuman = (state) => {
	return Object.values(state.conversations)
}

export { useMsgrContext }

const searchOne = (state, { searchText, convId, id }) => {
	const intes = state.interactions[convId]
	if (!intes) {
		return undefined
	}
	const inte = intes[id]
	return inte &&
		inte.type === AppMessageType.TypeUserMessage &&
		inte.payload.body &&
		inte.payload.body.toLowerCase().includes(searchText.toLowerCase())
		? inte
		: undefined
}

const messageToConvMapper = (conv) => (inte, messageIndex) => ({
	conversationId: conv.publicKey,
	id: inte.cid,
	conversationTitle: conv.displayName,
	messageIndex,
	interaction: inte,
})

export const useGetMessageSearchResultWithMetadata = (searchText) => {
	const ctx = useContext(MsgrContext)

	if (!searchText) {
		return []
	}

	// map all messages to conversation ID

	const conversations = listHuman(ctx)

	if (conversations.length > 30) {
		console.warn(
			'useGetMessageSearchResultWithMetadata: You have more than 30 conversations, this might take a while...',
		)
	}

	const messagesWithConv = flatten(
		conversations.map((conv) =>
			Object.values(ctx.interactions[conv.publicKey] || {}).map(messageToConvMapper(conv)),
		),
	)

	if (messagesWithConv.length > 200) {
		console.warn(
			'useGetMessageSearchResultWithMetadata: You have more than 200 messages, this might take a while...',
		)
	}

	// search messages for search match
	const messageIdsForConversationsWithMatch = messagesWithConv
		.map(({ conversationId, conversationTitle, messageIndex, id }) => ({
			id,
			conversationId,
			conversationTitle,
			messageIndex,
			message: searchOne(ctx, { searchText, convId: conversationId, id }),
		}))
		.filter(({ message }) => !!message && message.type === AppMessageType.TypeUserMessage)
		.sort(
			({ message: { sentDate: sentDateA } }, { message: { sentDate: sentDateB } }) =>
				sentDateB - sentDateA,
		)

	return messageIdsForConversationsWithMatch
}

export const useGetMessage = (id, convId) => {
	const ctx = useContext(MsgrContext)
	const intes = ctx.interactions[convId]
	if (!intes) {
		return undefined
	}
	return intes[id]
}

export const useAccountContactSearchResults = (searchText) => {
	const ctx = useContext(MsgrContext)
	if (!searchText) {
		return []
	}
	return Object.values(ctx.contacts).filter((contact) =>
		contact.displayName.toLowerCase().includes(searchText.toLowerCase()),
	)
}

export const useFirstConversationWithContact = (contactPk) => {
	const ctx = useContext(MsgrContext)
	const conversations = ctx.conversations
	const contact = ctx.contacts[contactPk]
	if (!contact) {
		return undefined
	}
	return conversations[contact.conversationPublicKey]
}

export const useAccount = () => {
	const ctx = useMsgrContext()
	return ctx.account
}

export const useContacts = () => {
	const ctx = useMsgrContext()
	return ctx.contacts
}

export const useContactsList = () => {
	const contacts = useContacts()
	return Object.values(contacts)
}

const ContactState = messengerpb.Contact.State

export const useIncomingContactRequests = () => {
	const contacts = useContactsList()
	return useMemo(() => contacts.filter((c) => c.state === ContactState.IncomingRequest), [contacts])
}

export const useOutgoingContactRequests = () => {
	const contacts = useContactsList()
	return useMemo(
		() =>
			contacts.filter((c) =>
				[ContactState.OutgoingRequestEnqueued, ContactState.OutgoingRequestSent].includes(c.state),
			),
		[contacts],
	)
}
