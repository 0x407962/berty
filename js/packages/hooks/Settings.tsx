import React from 'react'
import { settings } from '@berty-tech/store'
import { Provider as ReactReduxProvider, useSelector, useDispatch } from 'react-redux'
import { ActivityIndicator } from 'react-native'
import { PersistGate } from 'redux-persist/integration/react'
import * as Chat from './Chat'

export const Provider: React.FC = ({ children }) => {
	const store = settings.init()
	return (
		<ReactReduxProvider store={store}>
			<PersistGate loading={<ActivityIndicator size='large' />} persistor={store.persistor}>
				{children}
			</PersistGate>
		</ReactReduxProvider>
	)
}

export const useSettings = () => {
	const account = Chat.useAccount()
	return useSelector((state: settings.main.GlobalState) =>
		account ? settings.main.queries.get(state, { id: account.id }) : undefined,
	)
}

export const useToggleTracing = () => {
	const stgs = useSettings()
	const dispatch = useDispatch()
	if (!stgs) {
		return () => {}
	}
	return () => dispatch(settings.main.commands.toggleTracing({ id: stgs.id }))
}

export { settings as store }
