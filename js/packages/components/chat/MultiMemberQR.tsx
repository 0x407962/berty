import React from 'react'
import QRCode from 'react-native-qrcode-svg'
import { View } from 'react-native'
import { Button } from '@ui-kitten/components'
import { SafeAreaConsumer } from 'react-native-safe-area-context'

import { ScreenProps, useNavigation } from '@berty-tech/navigation'
import { useConversation } from '@berty-tech/store/hooks'
import { useStyles } from '@berty-tech/styles'

const _contentScaleFactor = 0.66

export const MultiMemberQR: React.FC<ScreenProps.Chat.MultiMemberQR> = ({
	route: {
		params: { convId },
	},
}) => {
	const [, { windowHeight, windowWidth }] = useStyles()
	const conv = useConversation(convId)
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
						value={conv.link}
					/>
					<Button style={[{ marginTop: 40 }]} onPress={goBack}>
						Go back
					</Button>
				</View>
			)}
		</SafeAreaConsumer>
	)
}
