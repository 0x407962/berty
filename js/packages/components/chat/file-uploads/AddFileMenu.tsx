import React from 'react'
import { TouchableOpacity, View, StyleSheet } from 'react-native'
import { useTranslation } from 'react-i18next'
import { Text, Icon } from '@ui-kitten/components'
import ImagePicker from 'react-native-image-crop-picker'

import { useStyles } from '@berty-tech/styles'
import { useClient } from '@berty-tech/store/hooks'

const ListItem: React.FC<{
	title: string
	onPress: () => void
	iconProps: {
		name: string
		fill: string
		height: number
		width: number
		pack?: string
	}
}> = ({ title, iconProps, onPress }) => {
	const [{ padding, margin }] = useStyles()

	return (
		<TouchableOpacity
			onPress={onPress}
			style={[
				padding.vertical.medium,
				padding.horizontal.large,
				{ flexDirection: 'row', alignItems: 'center' },
			]}
		>
			<Icon {...iconProps} />
			<Text style={[margin.left.large]}>{title}</Text>
		</TouchableOpacity>
	)
}

const amap = async <T extends any, C extends (value: T) => any>(arr: T[], cb: C) =>
	Promise.all(arr.map(cb))

export const AddFileMenu: React.FC<{ onClose: (medias: string[]) => void }> = ({ onClose }) => {
	console.log('AddFileMenu render')

	const client = useClient()
	const [{ color, border }] = useStyles()
	const { t } = useTranslation()

	const LIST_CONFIG = [
		{
			iconProps: {
				name: 'microphone',
				fill: '#C7C8D8',
				height: 40,
				width: 40,
				pack: 'custom',
			},
			title: t('chat.files.record-sound'),
			onPress: () => {},
		},
		{
			iconProps: {
				name: 'add-picture',
				fill: '#C7C8D8',
				height: 40,
				width: 40,
				pack: 'custom',
			},
			title: t('chat.files.media'),
			onPress: async () => {
				try {
					const res = await ImagePicker.openPicker({ multiple: true })
					console.log(res)
					const mediaCids = (
						await amap(res, async (doc) => {
							const stream = await client?.mediaPrepare({})
							await stream?.emit({
								info: { filename: doc.filename, mimeType: doc.mime, displayName: doc.filename },
								uri: doc.sourceURL || doc.path,
							})
							const reply = await stream?.stopAndRecv()
							return reply?.cid
						})
					).filter((cid) => !!cid)
					console.log(mediaCids)
					onClose(mediaCids)
				} catch (err) {
					if (err?.code !== 'E_PICKER_CANCELLED') {
						console.error('failed to add media', err)
					}
				}
			},
		},
		{
			iconProps: {
				name: 'bertyzzz',
				fill: 'white',
				height: 40,
				width: 40,
				pack: 'custom',
			},
			title: t('chat.files.emojis'),
			onPress: () => {},
		},
	]

	return (
		<View
			style={[
				StyleSheet.absoluteFill,
				{
					zIndex: 9,
					elevation: 9,
				},
			]}
		>
			<TouchableOpacity style={{ flex: 1 }} onPress={() => onClose([])}>
				<View></View>
			</TouchableOpacity>
			<View
				style={[
					{
						position: 'absolute',
						bottom: 100,
						left: 0,
						right: 0,
						width: '100%',
						backgroundColor: color.white,
					},
					border.radius.top.large,
					border.shadow.big,
				]}
			>
				{LIST_CONFIG.map((listItem) => (
					<ListItem {...listItem} key={listItem.title} />
				))}
			</View>
		</View>
	)
}
