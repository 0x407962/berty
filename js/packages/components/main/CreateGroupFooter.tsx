import React from 'react'
import { View, TouchableOpacity } from 'react-native'
import { Text, Icon } from '@ui-kitten/components'
import LinearGradient from 'react-native-linear-gradient'

import { useStyles } from '@berty-tech/styles'

type FooterCreateGroupProps = {
	title: string
	icon?: string
	action?: any
}

const useStylesCreateGroup = () => {
	const [{ padding, border, margin, text }] = useStyles()
	return {
		footerCreateGroup: [padding.horizontal.scale(60), margin.bottom.scale(40)],
		footerCreateGroupButton: border.radius.small,
		footerCreateGroupText: text.size.medium,
	}
}

export const FooterCreateGroup: React.FC<FooterCreateGroupProps> = ({ title, icon, action }) => {
	const [{ absolute, background, row, padding, color, text }] = useStyles()
	const _styles = useStylesCreateGroup()

	return (
		<>
			<LinearGradient
				style={[
					absolute.bottom,
					{ alignItems: 'center', justifyContent: 'center', height: '15%', width: '100%' },
				]}
				colors={['#ffffff00', '#ffffff80', '#ffffffc0', '#ffffffff']}
			/>
			<View style={[absolute.bottom, absolute.left, absolute.right, _styles.footerCreateGroup]}>
				<TouchableOpacity onPress={() => action()}>
					<View
						style={[
							background.light.blue,
							padding.horizontal.medium,
							padding.vertical.small,
							{
								flexDirection: 'row',
								justifyContent: 'center',
							},
							_styles.footerCreateGroupButton,
						]}
					>
						<View style={[row.item.justify]}>
							<Text
								style={[
									text.bold.medium,
									text.color.blue,
									text.align.center,
									_styles.footerCreateGroupText,
								]}
							>
								{title}
							</Text>
						</View>
						{icon && (
							<View style={[row.item.justify, padding.left.medium]}>
								<Icon name='arrow-forward-outline' width={25} height={25} fill={color.blue} />
							</View>
						)}
					</View>
				</TouchableOpacity>
			</View>
		</>
	)
}
