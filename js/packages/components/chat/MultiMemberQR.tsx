import React from 'react'
import QRCode from 'react-native-qrcode-svg'
import { View } from 'react-native'
import { Button } from 'react-native-ui-kitten'
import { SafeAreaConsumer } from 'react-native-safe-area-context'

import { ScreenProps, useNavigation } from '@berty-tech/navigation'
import { Messenger } from '@berty-tech/store/oldhooks'
import { useStyles } from '@berty-tech/styles'

const _contentScaleFactor = 0.66

export const MultiMemberQR: React.FC<ScreenProps.Chat.MultiMemberQR> = ({
	route: {
		params: { convId },
	},
}) => {
	const [, { windowHeight, windowWidth }] = useStyles()
	const conv = Messenger.useGetConversation(convId)
	const { goBack } = useNavigation()
	if (!conv) {
		return null
	}
	return (
		<SafeAreaConsumer>
			{(insets) => (
				<View
					style={[
						{
							paddingTop: insets?.top || 0,
							alignItems: 'center',
							height: '100%',
							justifyContent: 'center',
						},
					]}
				>
					<QRCode
						size={_contentScaleFactor * Math.min(windowHeight, windowWidth)}
						value={conv.shareableGroup}
					/>
					<Button style={[{ marginTop: 40 }]} onPress={goBack}>
						Go back
					</Button>
				</View>
			)}
		</SafeAreaConsumer>
	)
}
