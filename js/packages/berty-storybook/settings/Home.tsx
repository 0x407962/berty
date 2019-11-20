import React from 'react'
import { SafeAreaView, View, ScrollView, Image, TouchableOpacity, StyleSheet } from 'react-native'
import { Layout, Text, Icon } from 'react-native-ui-kitten'
import { colors, styles } from '../styles'
import { ButtonSetting, ButtonSettingRow } from '../shared-components/SettingsButtons'
import { RequestProps, UserProps } from '../shared-props/User'
import { Footer } from '../shared-components/Footer'
import { CircleAvatar } from '../shared-components/CircleAvatar'

//
// Home Vue
//

// Type

// Style
const _homeStyles = StyleSheet.create({
	homeAvatarBox: { width: 160, height: 180 },
	homeAvatar: { bottom: 37.5 },
	firstHeaderButton: {
		marginRight: 20,
		height: 90,
	},
	secondHeaderButton: {
		marginRight: 20,
		height: 90,
	},
	thirdHeaderButton: {
		height: 90,
	},
	headerNameText: {
		fontSize: 13,
	},
})

const HomeHeaderGroupButton: React.FC<{}> = () => (
	<View style={[styles.paddingRight, styles.paddingLeft]}>
		<ButtonSettingRow
			state={[
				{
					name: 'Updates',
					icon: 'arrow-upward-outline',
					color: colors.blue,
					style: _homeStyles.firstHeaderButton,
				},
				{
					name: 'Help',
					icon: 'question-mark-circle-outline',
					color: colors.red,
					style: _homeStyles.secondHeaderButton,
				},
				{
					name: 'Settings',
					icon: 'settings-2-outline',
					color: colors.blue,
					style: _homeStyles.thirdHeaderButton,
				},
			]}
		/>
	</View>
)

const HomeHeaderAvatar: React.FC<UserProps> = ({ avatarUri, name }) => (
	<View style={[styles.center, styles.marginTop]}>
		<View style={[_homeStyles.homeAvatarBox, styles.bgWhite, styles.borderRadius]}>
			<View style={[_homeStyles.homeAvatar]}>
				<CircleAvatar style={styles.centerItems} avatarUri={avatarUri} size={75} />
				<View style={[styles.center]}>
					<Text style={[styles.fontFamily, styles.littlePaddingTop, _homeStyles.headerNameText]}>
						{name}
					</Text>
				</View>
			</View>
		</View>
	</View>
)

const HomeHeader: React.FC<RequestProps> = ({ user }) => (
	<SafeAreaView
		style={[
			styles.bgBlue,
			styles.paddingLeft,
			styles.paddingRight,
			styles.borderBottomLeftRadius,
			styles.borderBottomRightRadius,
			styles.flex,
		]}
	>
		<TouchableOpacity style={[styles.end, styles.paddingRight]}>
			<Icon name='edit-outline' width={40} height={40} fill={colors.white} />
		</TouchableOpacity>
		<View style={[styles.alignVertical, styles.bigMarginBottom]}>
			<HomeHeaderAvatar {...user} />
		</View>
		<HomeHeaderGroupButton />
	</SafeAreaView>
)

const HomeBodySettings: React.FC<{}> = () => (
	<View style={[styles.flex, styles.paddingLeft, styles.paddingRight, styles.marginTop]}>
		<ButtonSetting
			name='Notifications'
			icon='bell-outline'
			state={{ value: 'Current', color: colors.white, bgColor: colors.blue }}
		/>
		<ButtonSetting name='Bluetooth' icon='bluetooth-outline' />
		<ButtonSetting name='Dark mode' icon='moon-outline' toggled />
		<ButtonSetting name='About Berty' icon='info-outline' />
		<ButtonSetting name='Devtools' icon='options-2-outline' />
		<ButtonSetting name='Devtools' icon='options-2-outline' />
		<ButtonSetting name='Devtools' icon='options-2-outline' />
	</View>
)

export const Home: React.FC<RequestProps> = ({ user }) => (
	<Layout style={[styles.flex, styles.bgWhite]}>
		<ScrollView contentContainerStyle={[styles.bgWhite]}>
			<HomeHeader user={user} />
			<HomeBodySettings />
		</ScrollView>
		<Footer
			left={{ icon: 'search-outline' }}
			center={{ icon: 'message-circle-outline' }}
			right={{ avatarUri: user.avatarUri, backgroundColor: colors.blue, size: 50, elemSize: 45 }}
		/>
	</Layout>
)
