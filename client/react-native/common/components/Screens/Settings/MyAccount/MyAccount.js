import React from 'react'
import { Screen, Menu, Header, Badge, Avatar } from '../../../Library'
import { colors } from '../../../../constants'
import { choosePicture } from '../../../../helpers/react-native-image-picker'
import I18n from 'i18next'
import { withNamespaces } from 'react-i18next'
import { withCurrentUser } from '../../../../utils/contact'
import RelayContext from '../../../../relay/RelayContext'
import { showMessage } from 'react-native-flash-message'

class MyAccountBase extends React.PureComponent {
  constructor (props) {
    super(props)

    this.state = {
      displayName: props.currentUser.displayName,
      uri: '',
    }
  }

  componentDidMount () {
    this.props.navigation.setParams({
      onSave: this.onSave,
    })
  }

  onSave = async () => {
    await this.props.context.mutations.contactUpdate({
      ...this.props.currentUser,
      displayName: this.state.displayName,
    })

    showMessage({
      message: I18n.t('contacts.info-updated'),
      type: 'info',
      position: 'top',
      icon: 'info',
    })

    this.props.navigation.goBack(null)
  }

  onChoosePicture = async event => this.setState(await choosePicture(event))

  render = () => {
    const { t, currentUser } = this.props
    const { displayName } = this.state

    return <Menu absolute>
      <Menu.Header
        icon={
          <Badge
            background={colors.blue}
            icon='camera'
            medium
            onPress={this.onChoosePicture}
          >
            <Avatar data={currentUser} size={78} />
          </Badge>
        }
      />
      <Menu.Section title={t('contacts.full-name')}>
        <Menu.Input
          value={displayName}
          onChangeText={displayName => this.setState({ displayName })}
        />
      </Menu.Section>
      <Menu.Section>
        <Menu.Item
          icon='trash-2'
          title={t('my-account.delete-my-account')}
          color={colors.error}
          onPress={() => console.error('delete my account: not implemented')}
        />
      </Menu.Section>
    </Menu>
  }
}

const MyAccountContent = withNamespaces()(withCurrentUser(MyAccountBase, { showOnlyLoaded: true }))

export default class MyAccount extends  React.Component {
  static navigationOptions = ({ navigation }) => {
    const onSave = navigation.getParam('onSave')
    return {
      tabBarVisible: false,
      header: (
        <Header
          navigation={navigation}
          title={I18n.t('my-account.title')}
          rightBtnIcon={'save'}
          onPressRightBtn={onSave}
          backBtn
        />
      ),
    }
  }

  render () {
    return <Screen>
      <RelayContext.Consumer>
        {context =>
          <MyAccountContent navigation={this.props.navigation} context={context} />
        }
      </RelayContext.Consumer>
    </Screen>
  }
}
