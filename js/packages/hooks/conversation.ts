import { useMemo } from 'react'
import { useDispatch, useSelector } from 'react-redux'
import { messenger } from '@berty-tech/store'
import { useAccount, useClient } from './account'

// conversations commands
export const useConversationGenerate = () => {
	const dispatch = useDispatch()
	return useMemo(() => () => dispatch(messenger.conversation.commands.generate()), [dispatch])
}

type UseConversationCreate = (kwargs: {
	members: messenger.contact.Entity[]
	name: string
}) => () => void

// multimember group
export const useConversationCreate: UseConversationCreate = ({ name, members }) => {
	const account = useAccount()
	const dispatch = useDispatch()
	return useMemo(() => {
		if (!account) {
			return () => {}
		}
		return () => {
			dispatch(messenger.conversation.commands.create({ accountId: account.id, name, members }))
		}
	}, [account, dispatch, members, name])
}

export const useConversationDelete = () => {
	const dispatch = useDispatch()
	return useMemo(
		() => (payload: messenger.conversation.Command.Delete) =>
			dispatch(messenger.conversation.commands.delete(payload)),
		[dispatch],
	)
}

export const useStartReadConversation = (id: messenger.conversation.Entity['id']) => {
	const dispatch = useDispatch()
	return useMemo(() => () => dispatch(messenger.conversation.commands.startRead(id)), [
		dispatch,
		id,
	])
}

export const useStopReadConversation = (id: messenger.conversation.Entity['id']) => {
	const dispatch = useDispatch()
	return useMemo(() => () => dispatch(messenger.conversation.commands.stopRead(id)), [dispatch, id])
}

// conversation queries
export const useConversationList = () => {
	const client = useClient()
	// TODO: handle multiple devices
	const list = useSelector((state: messenger.conversation.GlobalState) =>
		client
			? messenger.conversation.queries
					.list(state)
					.filter(
						(conv) =>
							conv.kind === 'fake' ||
							Object.keys(conv.membersDevices).filter((m) => m !== client.accountPk).length > 0,
					)
			: [],
	)
	return list
}

export const useConversationLength = () => {
	return useConversationList().length
}

export const useGetConversation = (id: string): messenger.conversation.Entity | undefined => {
	const conversation = useSelector((state: messenger.conversation.GlobalState) =>
		messenger.conversation.queries.get(state, { id }),
	)
	return conversation
}

export const useOneToOneConversationContact = (
	id: string,
): messenger.contact.Entity | undefined => {
	const conversation = useGetConversation(id)
	return useSelector(
		(state: messenger.contact.GlobalState) =>
			(conversation?.kind === messenger.conversation.ConversationKind.OneToOne &&
				messenger.contact.queries.get(state, { id: conversation.contactId })) ||
			undefined,
	)
}
