import React, { useMemo } from 'react'
import { chat } from '@berty-tech/store'
import { Provider as ReactReduxProvider, useDispatch, useSelector } from 'react-redux'
import DevMenu from 'react-native-dev-menu'
import { Clipboard } from 'react-native'

export const Recorder: React.FC = ({ children }) => {
	React.useEffect(() => {
		DevMenu.addItem('(Chat) Start Test Recorder', () => {
			chat.recorder.start()
		})
		DevMenu.addItem('(Chat) Copy Test And Stop Recoder', () => {
			Clipboard.setString(
				chat.recorder
					.createTest()
					.replace(
						'/* import reducer from YOUR_REDUCER_LOCATION_HERE */',
						"import * as chat from '..'\nconst { reducer } = chat.init()",
					),
			)
			chat.recorder.stop()
		})
	})

	return null
}

export const Provider: React.FC = ({ children }) => {
	return <ReactReduxProvider store={chat.init()}>{children}</ReactReduxProvider>
}

// account commands
export const useAccountGenerate = () => {
	const dispatch = useDispatch()
	return useMemo(() => () => dispatch(chat.account.commands.generate()), [dispatch])
}

export const useAccountCreate = () => {
	const dispatch = useDispatch()
	return useMemo(
		() => (payload: chat.account.Command.Create) => dispatch(chat.account.commands.create(payload)),
		[dispatch],
	)
}

// account queries
export const useAccountList = () => {
	const list = useSelector((state: chat.account.GlobalState) =>
		chat.account.queries.list(state, {}),
	)
	return list
}

export const useAccountLength = () => {
	return useAccountList().length
}

export const useAccount = () => {
	// TODO: replace by selected account
	const accounts = useAccountList()
	const len = useAccountLength()
	return len > 0 ? accounts[0] : null
}

export const useAccountContactRequestReference = () => {
	const account = useAccount()
	return account?.contactRequestReference
}

export const useAccountContactRequestEnabled = () => {
	const ref = useAccountContactRequestReference()
	return !!ref
}

export const useAccountSendContactRequest = () => {
	const dispatch = useDispatch()
	const account = useAccount()
	if (!account) {
		return () => {}
	}
	return (reference: string) => {
		dispatch(
			chat.account.commands.sendContactRequest({
				id: account.id,
				otherReference: reference,
			}),
		)
	}
}

// requests queries
export const useIncomingContactRequests = () => {
	return useSelector((state: chat.incomingContactRequest.GlobalState) =>
		chat.incomingContactRequest.queries.list(state, {}),
	)
}

export const useOutgoingContactRequests = () => {
	return useSelector((state: chat.outgoingContactRequest.GlobalState) =>
		chat.outgoingContactRequest.queries.list(state, {}),
	)
}

export const useAccountIncomingContactRequests = () => {
	const account = useAccount()
	const incomingContactRequests = useIncomingContactRequests()
	if (!account) {
		return []
	}
	return incomingContactRequests.filter((req) => req.accountId === account.id)
}

export const useAccountOutgoingContactRequests = () => {
	const account = useAccount()
	const incomingContactRequests = useOutgoingContactRequests()
	if (!account) {
		return []
	}
	return incomingContactRequests.filter((req) => req.accountId === account.id)
}

export const useAccountAcceptContactRequest = () => {
	const dispatch = useDispatch()
	return ({ id }: { id: string }) =>
		dispatch(
			chat.incomingContactRequest.commands.accept({
				aggregateId: id,
			}),
		)
}

export const useAccountDiscardContactRequest = () => {
	const dispatch = useDispatch()
	return ({ id }: { id: string }) =>
		dispatch(
			chat.incomingContactRequest.commands.discard({
				aggregateId: id,
			}),
		)
}
