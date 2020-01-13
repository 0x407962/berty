import React from 'react'
import {
	View,
	SafeAreaView,
	ScrollView,
	TouchableOpacity,
	StyleSheet,
	Dimensions,
	TouchableWithoutFeedback,
} from 'react-native'
import { Layout, Text, Icon } from 'react-native-ui-kitten'
import { styles, colors } from '@berty-tech/styles'
import { BlurView } from '@react-native-community/blur'
import { SDTSModalComponent } from '../shared-components/SDTSModalComponent'
import { CircleAvatar } from '../shared-components/CircleAvatar'
import { useNavigation } from '@berty-tech/berty-navigation'

const _stylesList = StyleSheet.create({
	tinyAvatar: {
		position: 'absolute',
		top: -32.5,
	},
	tinyCard: {
		margin: 16,
		marginTop: 16 + 26,
		padding: 16,
		paddingTop: 16 + 26,
		width: 121,
		height: 177,
		borderRadius: 20,
		backgroundColor: colors.white,
		alignItems: 'center',
	},
	tinyAcceptButton: {
		paddingHorizontal: 8,
		paddingVertical: 4,
		borderRadius: 4,
		marginHorizontal: 4,
	},
	tinyDiscardButton: {
		paddingHorizontal: 4,
		paddingVertical: 4,
		borderRadius: 4,
		marginHorizontal: 4,
	},
	addContactItem: {
		height: 115,
		width: 150,
	},
	addContactItemText: {
		width: 75,
	},
})

const RequestsItem: React.FC<{}> = () => {
	const navigation = useNavigation()
	return (
		<TouchableOpacity
			style={[_stylesList.tinyCard, styles.shadow, styles.col]}
			onPress={navigation.navigate.main.requestSent}
		>
			<CircleAvatar
				style={_stylesList.tinyAvatar}
				avatarUri='https://s3.amazonaws.com/uifaces/faces/twitter/msveet/128.jpg'
				size={65}
				diffSize={8}
			/>
			<Text numberOfLines={1} style={[styles.center, styles.textCenter, styles.flex]}>
				Gjdgfhnd
			</Text>
			<Text
				category='c1'
				style={[styles.paddingVertical, styles.textCenter, styles.textTiny, styles.textGrey]}
			>
				Sent 3 days ago
			</Text>
			<View style={[styles.row]}>
				<TouchableOpacity
					style={[_stylesList.tinyDiscardButton, styles.border, styles.justifyContent]}
				>
					<Icon name='close-outline' width={15} height={15} fill={colors.grey} />
				</TouchableOpacity>
				<TouchableOpacity
					style={[
						_stylesList.tinyAcceptButton,
						styles.bgLightGreen,
						styles.row,
						styles.alignItems,
						styles.justifyContent,
					]}
				>
					<Icon name='checkmark-outline' width={15} height={15} fill={colors.green} />
					<Text style={[styles.textTiny, styles.textGreen]}>Resend</Text>
				</TouchableOpacity>
			</View>
		</TouchableOpacity>
	)
}

const Requests: React.FC<{}> = () => (
	<SafeAreaView>
		<View style={[styles.paddingVertical]}>
			<ScrollView horizontal showsHorizontalScrollIndicator={false}>
				<RequestsItem />
				<RequestsItem />
				<RequestsItem />
				<RequestsItem />
				<RequestsItem />
				<RequestsItem />
				<RequestsItem />
				<RequestsItem />
				<RequestsItem />
				<RequestsItem />
				<RequestsItem />
				<RequestsItem />
				<RequestsItem />
				<RequestsItem />
				<RequestsItem />
			</ScrollView>
		</View>
	</SafeAreaView>
)

const AddContact: React.FC<{}> = () => {
	const navigation = useNavigation()
	return (
		<View style={[styles.paddingVertical]}>
			<View style={[styles.row, styles.spaceAround]}>
				<TouchableOpacity
					style={[
						styles.col,
						styles.padding,
						styles.borderRadius,
						styles.bgRed,
						_stylesList.addContactItem,
					]}
					onPress={navigation.navigate.main.scan}
				>
					<View style={[styles.row, styles.spaceBetween]}>
						<View />
						<Icon name='image-outline' height={50} width={50} fill={colors.white} />
					</View>
					<View style={[styles.row, styles.spaceBetween]}>
						<Text numberOfLines={2} style={[styles.textWhite, _stylesList.addContactItemText]}>
							Scan QR code
						</Text>
						<View />
					</View>
				</TouchableOpacity>
				<TouchableOpacity
					style={[
						styles.col,
						styles.padding,
						styles.borderRadius,
						styles.bgBlue,
						_stylesList.addContactItem,
					]}
					onPress={navigation.navigate.settings.myBertyId}
				>
					<View style={[styles.row, styles.spaceBetween]}>
						<View />
						<Icon name='person-outline' height={50} width={50} fill={colors.white} />
					</View>
					<View style={[styles.row, styles.spaceBetween]}>
						<Text numberOfLines={2} style={[styles.textWhite, _stylesList.addContactItemText]}>
							Share my Berty ID
						</Text>
						<View />
					</View>
				</TouchableOpacity>
			</View>
		</View>
	)
}

const NewGroup: React.FC<{}> = () => <View />

const Screen = Dimensions.get('window')

export const ListModal: React.FC<{}> = () => {
	const firstNotToggledPoint = Screen.height - 283 + 16 + 35
	const firstToggledPoint = firstNotToggledPoint

	const secondNotToggledPoint = firstToggledPoint - 200
	const secondToggledPoint = secondNotToggledPoint - 163 + 20

	const thirdNotToggledPoint = secondToggledPoint - 200
	const thirdToggledPoint = thirdNotToggledPoint - 283 + 20
	const navigation = useNavigation()

	return (
		<>
			<TouchableWithoutFeedback
				onPress={navigation.goBack}
				style={[styles.test, StyleSheet.absoluteFill]}
			>
				<BlurView style={StyleSheet.absoluteFill} blurType='light' />
			</TouchableWithoutFeedback>
			<SafeAreaView style={[styles.absolute, styles.bottom, styles.right, styles.left]}>
				<SDTSModalComponent
					rows={[
						{
							toggledPoint: firstToggledPoint,
							notToggledPoint: firstNotToggledPoint,
							title: 'New group',
							icon: 'people-outline',
							iconColor: colors.black,
							dragEnabled: false,
							headerAction: navigation.navigate.main.createGroup2,
						},
						{
							toggledPoint: secondToggledPoint,
							notToggledPoint: secondNotToggledPoint,
							title: 'Add contact',
							icon: 'person-add-outline',
							iconColor: colors.black,
						},
						{
							toggledPoint: thirdToggledPoint,
							notToggledPoint: thirdNotToggledPoint,
							title: 'Requests sent',
							icon: 'paper-plane-outline',
							iconColor: colors.black,
						},
					]}
				>
					<NewGroup />
					<AddContact />
					<Requests />
				</SDTSModalComponent>
			</SafeAreaView>
		</>
	)
}
