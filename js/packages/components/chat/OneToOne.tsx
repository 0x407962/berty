import React, { useState, useRef, useMemo, useEffect } from 'react'
import {
	TouchableOpacity,
	View,
	StyleSheet,
	FlatList,
	ActivityIndicator,
	KeyboardAvoidingView,
	Text as TextNative,
} from 'react-native'
import { Text, Icon } from 'react-native-ui-kitten'
import { CommonActions } from '@react-navigation/native'

import { useStyles } from '@berty-tech/styles'
import { useNavigation, ScreenProps, Routes } from '@berty-tech/navigation'
import * as api from '@berty-tech/api/index.pb'
import { messenger as messengerpb } from '@berty-tech/api/index.js'
import {
	useMsgrContext,
	useConversation,
	useContact,
	useReadEffect,
	useSortedConvInteractions,
	usePersistentOptions,
} from '@berty-tech/store/hooks'

import { ProceduralCircleAvatar } from '../shared-components/ProceduralCircleAvatar'
import {
	Message,
	MessageInvitationWrapper,
	MessageInvitationButton,
} from './shared-components/Message'
import BlurView from '../shared-components/BlurView'
import messengerMethodsHooks from '@berty-tech/store/methods'

import moment from 'moment'

import { ChatFooter, ChatDate } from './shared-components/Chat'
import { SwipeNavRecognizer } from '../shared-components/SwipeNavRecognizer'
import Logo from '../main/1_berty_picto.svg'
import Avatar from '../modals/Buck_Berty_Icon_Card.svg'

//
// Chat
//

// Styles
const useStylesChat = () => {
	const [{ flex, text }] = useStyles()
	return {
		headerName: flex.large,
		headerNameText: text.size.scale(20),
	}
}

const useStylesAddBetabot = () => {
	const [{ width, border, padding, margin }] = useStyles()
	return {
		skipButton: [
			border.color.light.grey,
			border.scale(2),
			border.radius.small,
			margin.top.scale(15),
			padding.left.small,
			padding.right.medium,
			padding.top.small,
			padding.bottom.small,
			width(120),
		],
		addButton: [
			border.color.light.blue,
			border.scale(2),
			border.radius.small,
			margin.top.scale(15),
			padding.left.small,
			padding.right.medium,
			padding.top.small,
			padding.bottom.small,
			width(120),
		],
	}
}

const CenteredActivityIndicator: React.FC = (props: ActivityIndicator['props']) => {
	const { children, ...propsToPass } = props
	return (
		<View style={{ width: '100%', height: '100%', justifyContent: 'center', alignItems: 'center' }}>
			<ActivityIndicator {...propsToPass} />
		</View>
	)
}

export const ChatHeader: React.FC<{ convPk: any }> = ({ convPk }) => {
	const { navigate, goBack } = useNavigation()
	const conv = useConversation(convPk)
	const contact = useContact(conv?.contactPublicKey || null)

	const _styles = useStylesChat()
	const [
		{ row, padding, column, margin, text, flex, opacity, color, border, background },
		{ scaleHeight },
	] = useStyles()

	const persistOpts = usePersistentOptions()
	const isBetabot =
		persistOpts && conv.contactPublicKey.toString() === persistOpts.betabot.convPk.toString()

	// const lastDate = Messenger.useGetDateLastContactMessage(convPk)
	// const lastDate = new Date()
	// const debugGroup = Settings.useDebugGroup({ pk: conv?.pk || '' })
	// const main = Settings?.useSettings()
	// const state = main?.debugGroup?.state

	// useEffect(() => {
	// 	if (!state) {
	// 		debugGroup()
	// 	}
	// })

	// useEffect(() => {
	// 	const interval = setInterval(() => {
	// 		debugGroup()
	// 	}, 10000)
	// 	return () => clearInterval(interval)
	// }, [debugGroup])

	if (!conv || !contact) {
		goBack()
		console.warn('OneToOne: no conv', conv, contact)
		return <CenteredActivityIndicator />
	}
	const title = conv.fake ? `FAKE - ${contact.displayName}` : contact?.displayName || ''
	return (
		<View style={{ position: 'absolute', top: 0, left: 0, right: 0 }}>
			<BlurView
				blurType='light'
				blurAmount={30}
				style={{ position: 'absolute', bottom: 0, top: 0, left: 0, right: 0 }}
			/>
			<View
				style={[
					padding.right.medium,
					padding.left.tiny,
					{
						alignItems: 'center',
						flexDirection: 'row',
						marginTop: 50 * scaleHeight,
						paddingBottom: 20 * scaleHeight,
					},
				]}
			>
				<TouchableOpacity
					style={[flex.tiny, { justifyContent: 'center', alignItems: 'center' }]}
					onPress={goBack}
				>
					<Icon name='arrow-back-outline' width={25} height={25} fill={color.black} />
				</TouchableOpacity>
				<View
					style={[
						flex.huge,
						column.justify,
						row.item.justify,
						margin.top.small,
						_styles.headerName,
					]}
				>
					<View style={[{ flexDirection: 'row', justifyContent: 'center', alignItems: 'center' }]}>
						<TouchableOpacity
						// onPress={() => debugGroup()}
						>
							<Text
								numberOfLines={1}
								style={[text.align.center, text.bold.medium, _styles.headerNameText]}
							>
								{title}
							</Text>
						</TouchableOpacity>
						{/* {state === 'error' && (
							<Icon name='close-outline' width={14} height={14} fill={color.red} />
						)} */}
						{/* {state === 'done' ? (
							<View
								style={[
									width(14),
									height(14),
									border.radius.scale(7),
									margin.left.large,
									{
										backgroundColor: main?.debugGroup?.peerIds?.length ? color.green : color.red,
									},
								]}
							/>
						) : (
							<ActivityIndicator size='small' style={[margin.left.large]} />
						)} */}
					</View>
					{/* {lastDate && (
						<Text numberOfLines={1} style={[text.size.small, text.color.grey, text.align.center]}>
							Last seen <FromNow date={lastDate} />
						</Text>
					)} */}
				</View>
				<View style={[flex.tiny, row.fill, { alignItems: 'center' }]}>
					<TouchableOpacity
						activeOpacity={contact ? 0.2 : 0.5}
						style={[flex.tiny, row.item.justify, !contact ? opacity(0.5) : null]}
						onPress={() => navigate.chat.oneToOneSettings({ convId: convPk })}
					>
						{!isBetabot ? (
							<ProceduralCircleAvatar size={45} diffSize={9} seed={conv.contactPublicKey} />
						) : (
							<View
								style={[
									border.radius.scale(25),
									border.shadow.medium,
									background.white,
									margin.right.small,
									{
										justifyContent: 'center',
										alignItems: 'center',
										display: 'flex',
										width: 40,
										height: 40,
										alignSelf: 'center',
									},
								]}
							>
								<Logo width={25} height={25} style={{ right: -1, top: -1 }} />
							</View>
						)}
					</TouchableOpacity>
				</View>
			</View>
		</View>
	)
}

const ContactInitiatedWrapper: React.FC<{ children: any }> = ({ children }) => {
	const [
		{ padding, border, flex, margin, width, background, height, minWidth },
		{ scaleSize },
	] = useStyles()
	const logoDiameter = 28
	const diffSize = 6
	return (
		<View
			style={[
				{ backgroundColor: '#EDEEF8' },
				padding.bottom.tiny,
				padding.horizontal.medium,
				margin.top.scale(logoDiameter + 10),
				border.radius.scale(10),
				flex.justify.center,
				minWidth(0),
				{ shadowOpacity: 0.1, shadowRadius: 4, shadowOffset: { width: 0, height: 2.5 } },
			]}
		>
			<View
				style={{
					transform: [{ translateY: -logoDiameter * 1.15 * scaleSize }],
					alignSelf: 'center',
					marginBottom: -logoDiameter * scaleSize, // compensate for transformed logo
				}}
			>
				<View
					style={[
						flex.align.center,
						flex.justify.center,
						width(logoDiameter + diffSize * scaleSize),
						height(logoDiameter + diffSize * scaleSize),
						background.white,
						border.radius.scale((logoDiameter + diffSize) / 2),
						{
							borderWidth: 2,
							borderColor: 'rgba(215, 217, 239, 1)',
						},
					]}
				>
					<Logo
						width={scaleSize * logoDiameter - diffSize}
						height={scaleSize * logoDiameter - diffSize}
						style={[margin.left.tiny]} // nudge logo to appear centered
					/>
				</View>
			</View>
			{children}
		</View>
	)
}

const InfosContactState: React.FC<{ state: any }> = ({ state }) => {
	const [{ text, border, padding, margin }] = useStyles()
	const textColor = '#4E58BF'
	return (
		<View
			style={[
				border.radius.medium,
				padding.tiny,
				padding.horizontal.medium,
				margin.top.tiny,
				{
					flexDirection: 'row',
					justifyContent: 'space-evenly',
					alignItems: 'center',
					backgroundColor: '#EDEEF8',
				},
			]}
		>
			<Icon name='info-outline' fill={textColor} width={25} height={25} />
			<Text style={[text.italic, text.bold.small, padding.left.small, { color: textColor }]}>
				{state}
			</Text>
		</View>
	)
}

const ContactRequestBox: React.FC<{ contact: any; isAccepted: boolean }> = ({
	contact,
	isAccepted,
}) => {
	const { publicKey, displayName } = contact
	const [{ row, flex, text, margin, color }] = useStyles()
	const [acceptDisabled, setAcceptDisabled] = useState<boolean>(false)

	const { refresh: accept } = messengerMethodsHooks.useContactAccept()
	const decline: any = () => {}

	useEffect(() => {
		if (isAccepted) {
			setAcceptDisabled(true)
		}
	}, [isAccepted])

	return (
		<View>
			<View style={[row.left, flex.align.center, flex.justify.center]}>
				<TextNative
					style={[
						text.color.black,
						text.size.scale(15),
						text.bold.medium,
						{ fontFamily: 'Open Sans' },
					]}
				>
					CONTACT REQUEST
				</TextNative>
			</View>
			<View style={[margin.top.small, flex.align.center, flex.justify.center]}>
				<ProceduralCircleAvatar seed={publicKey || '42'} size={40} style={[margin.bottom.small]} />
				<TextNative
					style={[
						text.color.black,
						text.size.scale(13),
						text.bold.small,
						margin.bottom.small,
						{ fontFamily: 'Open Sans' },
					]}
				>
					{displayName}
				</TextNative>
			</View>
			<View style={[row.center, flex.justify.spaceEvenly, flex.align.center, margin.top.medium]}>
				<MessageInvitationButton
					onPress={() => decline()}
					activeOpacity={!acceptDisabled ? 0.2 : 1}
					icon='close-outline'
					color={color.grey}
					title='REFUSE'
					backgroundColor={color.white}
					styleOpacity={0.6}
					disabled
				/>
				<MessageInvitationButton
					onPress={() => accept({ publicKey })}
					activeOpacity={!acceptDisabled ? 0.2 : 1}
					icon='checkmark-outline'
					color={!acceptDisabled ? color.blue : color.green}
					title={!acceptDisabled ? 'ACCEPT' : 'ACCEPTED'}
					backgroundColor={!acceptDisabled ? color.light.blue : color.light.green}
					styleOpacity={acceptDisabled ? 0.6 : undefined}
					disabled={acceptDisabled}
				/>
			</View>
		</View>
	)
}

export const AddBetabotBox = () => {
	const [
		{ row, text, margin, color, padding, background, border, opacity },
		{ scaleHeight },
	] = useStyles()
	const _styles = useStylesAddBetabot()
	const { setPersistentOption, persistentOptions } = useMsgrContext()

	return (
		<View
			style={[
				{
					justifyContent: 'center',
					alignItems: 'center',
				},
				padding.medium,
			]}
		>
			<View
				style={[
					{
						width: 80 * scaleHeight,
						height: 80 * scaleHeight,
						backgroundColor: 'white',
						justifyContent: 'center',
						alignItems: 'center',
						position: 'relative',
						top: 10,
						zIndex: 1,
						shadowOpacity: 0.1,
						shadowRadius: 5,
						shadowOffset: { width: 0, height: 3 },
					},
					background.white,
					border.radius.scale(60),
				]}
			>
				<Avatar width={80 * scaleHeight} height={80 * scaleHeight} />
			</View>
			<View
				style={[
					background.white,
					padding.horizontal.medium,
					padding.bottom.medium,
					border.radius.large,
					border.shadow.huge,
					{ width: '100%' },
				]}
			>
				<View style={[padding.top.scale(35)]}>
					<Icon
						name='info-outline'
						fill={color.blue}
						width={60 * scaleHeight}
						height={60 * scaleHeight}
						style={[row.item.justify, padding.top.large]}
					/>
					<TextNative
						style={[
							text.align.center,
							padding.top.small,
							text.size.large,
							text.bold.medium,
							text.color.black,
							{ fontFamily: 'Open Sans' },
						]}
					>
						ADD BETA BOT?
					</TextNative>
					<Text style={[text.align.center, padding.top.scale(20), padding.horizontal.medium]}>
						<TextNative
							style={[
								text.bold.small,
								text.size.medium,
								text.color.black,
								{ fontFamily: 'Open Sans' },
							]}
						>
							You don't have any contacts yet would you like to add the
						</TextNative>
						<TextNative
							style={[
								text.bold.medium,
								text.size.medium,
								text.color.black,
								{ fontFamily: 'Open Sans' },
							]}
						>
							{' '}
							Beta Bot
						</TextNative>
						<TextNative
							style={[
								text.bold.small,
								text.color.black,
								text.size.medium,
								{ fontFamily: 'Open Sans' },
							]}
						>
							{' '}
							to discover and test conversations?
						</TextNative>
					</Text>
				</View>
				<View style={[row.center, padding.top.medium]}>
					<TouchableOpacity
						style={[row.fill, margin.bottom.medium, opacity(0.5), _styles.skipButton]}
					>
						<Icon name='close' width={30} height={30} fill={color.grey} style={row.item.justify} />
						<TextNative
							style={[
								text.color.grey,
								padding.left.small,
								row.item.justify,
								text.size.scale(16),
								text.bold.medium,
								{ fontFamily: 'Open Sans' },
							]}
						>
							SKIP
						</TextNative>
					</TouchableOpacity>
					<TouchableOpacity
						style={[row.fill, margin.bottom.medium, background.light.blue, _styles.addButton]}
						onPress={() => {
							setPersistentOption('betabot', {
								added: true,
								convPk: persistentOptions.betabot.convPk,
							})
						}}
					>
						<Icon
							name='checkmark-outline'
							width={30}
							height={30}
							fill={color.blue}
							style={row.item.justify}
						/>
						<TextNative
							style={[
								text.color.blue,
								padding.left.small,
								row.item.justify,
								text.size.scale(16),
								text.bold.medium,
							]}
						>
							ADD !
						</TextNative>
					</TouchableOpacity>
				</View>
			</View>
		</View>
	)
}

const InfosChat: React.FC<api.berty.messenger.v1.IConversation & any> = ({
	createdDate: createdDateStr,
	publicKey,
	isBetabot,
	isBetabotAdded,
}) => {
	const [{ flex, text, padding, margin }] = useStyles()
	const createdDate =
		(createdDateStr && parseInt((createdDateStr as unknown) as string, 10)) || Date.now()
	const ctx = useMsgrContext()
	const contact =
		Object.values(ctx.contacts).find((c: any) => c.conversationPublicKey === publicKey) || null
	const isAccepted = contact.state === messengerpb.Contact.State.Accepted
	const isIncoming = contact.state === messengerpb.Contact.State.IncomingRequest
	const textColor = '#4E58BF'

	return (
		<View style={[padding.medium, flex.align.center]}>
			<ChatDate date={createdDate} />
			{!isIncoming ? (
				<>
					{isBetabot && !isBetabotAdded ? (
						<AddBetabotBox />
					) : (
						<ContactInitiatedWrapper>
							<Text style={[text.align.center, text.italic, { color: textColor }]}>
								{isAccepted ? '👋 Berty connection confirmed! 🎉' : 'Request Sent'}
							</Text>
						</ContactInitiatedWrapper>
					)}
				</>
			) : (
				<MessageInvitationWrapper>
					<ContactRequestBox contact={contact} isAccepted={isAccepted} />
				</MessageInvitationWrapper>
			)}
			{!isAccepted && contact.state !== messengerpb.Contact.State.Undefined && (
				<>
					<View style={[flex.align.center]}>
						<Text style={[text.size.tiny, text.color.grey, margin.top.tiny]}>
							{Date.now() - new Date(createdDate).getTime() > 86400000
								? moment(createdDate).format('DD/MM/YYYY')
								: moment(createdDate).format('hh:mm')}
						</Text>
					</View>
					<InfosContactState
						state={
							isIncoming
								? 'Accept this request to reply!'
								: 'Waiting for your contact reply to connect!'
						}
					/>
				</>
			)}
		</View>
	)
}

const MessageList: React.FC<{ convPk: string; scrollToMessage?: string }> = ({
	convPk,
	scrollToMessage,
}) => {
	const ctx: any = useMsgrContext()
	const conv = ctx.conversations[convPk]
	const messages = useSortedConvInteractions(convPk).filter(
		(msg) =>
			msg.type === messengerpb.AppMessage.Type.TypeUserMessage ||
			msg.type === messengerpb.AppMessage.Type.TypeGroupInvitation,
	)

	if (conv.replyOptions !== null) {
		messages.push(conv.replyOptions)
	}

	const flatListRef: any = useRef(null)

	const onScrollToIndexFailed = () => {
		// Not sure why this happens (something to do with item/screen dimensions I think)
		flatListRef.current?.scrollToIndex({ index: 0 })
	}

	const initialScrollIndex = React.useMemo(() => {
		if (scrollToMessage) {
			for (let i = 0; i < messages.length; i++) {
				if (messages[i].cid === scrollToMessage) {
					return i
				}
			}
		}
	}, [messages, scrollToMessage])

	const [{ overflow, row, flex }, { scaleHeight }] = useStyles()

	const items: any = React.useMemo(() => {
		return messages?.reverse() || []
	}, [messages])

	const persistOpts = usePersistentOptions()
	const isBetabot =
		persistOpts && conv.contactPublicKey.toString() === persistOpts.betabot.convPk.toString()
	const isBetabotAdded = persistOpts && persistOpts.betabot.added

	return (
		<FlatList
			initialScrollIndex={initialScrollIndex}
			onScrollToIndexFailed={onScrollToIndexFailed}
			style={[overflow, row.item.fill, flex.tiny, { marginTop: 105 * scaleHeight }]}
			ref={flatListRef}
			keyboardDismissMode='on-drag'
			data={items}
			inverted
			keyExtractor={(item: any) => item.cid}
			ListFooterComponent={
				<InfosChat {...conv} isBetabot={isBetabot} isBetabotAdded={isBetabotAdded} />
			}
			renderItem={({ item, index }: { item: any; index: number }) => {
				if (isBetabot && !isBetabotAdded) {
					return null
				}
				return (
					<Message
						id={item.cid}
						convKind={messengerpb.Conversation.Type.ContactType}
						convPK={conv.publicKey}
						previousMessageId={index > 0 ? items[index + 1]?.cid : ''}
						nextMessageId={items[index - 1]?.cid || ''}
					/>
				)
			}}
		/>
	)
}

export const OneToOne: React.FC<ScreenProps.Chat.OneToOne> = ({ route: { params } }) => {
	const [inputIsFocused, setInputFocus] = useState(false)
	const [{ flex, background }] = useStyles()
	useReadEffect(params?.convId, 1000)
	const { dispatch } = useNavigation()

	const ctx: any = useMsgrContext()
	const conv = ctx.conversations[params?.convId]
	const contact =
		Object.values(ctx.contacts).find((c: any) => c.conversationPublicKey === conv?.publicKey) ||
		null
	const isIncoming = contact.state === messengerpb.Contact.State.IncomingRequest
	const persistOpts = usePersistentOptions()
	const isBetabotAdded = persistOpts && persistOpts.betabot.added

	return (
		<View style={[StyleSheet.absoluteFill, background.white, { flex: 1 }]}>
			<SwipeNavRecognizer
				onSwipeLeft={() =>
					dispatch(
						CommonActions.navigate({
							name: Routes.Chat.OneToOneSettings,
							params: { convId: params?.convId },
						}),
					)
				}
			>
				<KeyboardAvoidingView
					style={[flex.tiny, { justifyContent: 'flex-start' }]}
					behavior='padding'
				>
					<MessageList convPk={params?.convId} scrollToMessage={params?.scrollToMessage || '0'} />
					<ChatFooter
						convPk={params?.convId}
						isFocused={inputIsFocused}
						setFocus={setInputFocus}
						disabled={isIncoming || (isBetabotAdded ? false : true)}
					/>
					<ChatHeader convPk={params?.convId || ''} />
				</KeyboardAvoidingView>
			</SwipeNavRecognizer>
		</View>
	)
}
