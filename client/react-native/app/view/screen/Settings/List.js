import React, { PureComponent } from 'react'

import { Menu, Screen, Avatar, Header, Loader } from '@berty/component'
import { colors } from '@berty/common/constants'
import { fragments } from '@berty/graphql'
import { UpdateContext, installUpdate } from '@berty/update'
import { withCurrentUser } from '@berty/relay/utils/contact'
import { withNamespaces } from 'react-i18next'
import { contact } from '@berty/relay/utils'
import I18n from 'i18next'

class List extends PureComponent {
  static navigationOptions = ({ navigation }) => ({
    header: (
      <Header
        navigation={navigation}
        title={I18n.t('settings.title')}
        titleIcon='settings'
      />
    ),
    tabBarVisible: true,
  })

  static Menu = fragments.Contact(
    ({ navigation, data, availableUpdate, t }) => {
      const { id, displayName, overrideDisplayName } = data

      return (
        <Menu absolute>
          <Menu.Header
            icon={<Avatar data={{ id }} size={78} />}
            title={overrideDisplayName || displayName}
          />
          <Menu.Section>
            <Menu.Item
              icon='user'
              title={t('settings.my-account')}
              onPress={() => navigation.navigate('settings/my-account', {})}
            />
            <Menu.Item
              icon='share'
              title={t('settings.my-account-share')}
              onPress={() =>
                navigation.navigate('modal/contacts/card', {
                  ...data,
                  id: contact.getCoreID(id),
                  self: true,
                })
              }
            />
            {availableUpdate ? (
              <Menu.Item
                icon='arrow-up-circle'
                title={t('settings.update-available')}
                onPress={() => installUpdate(availableUpdate)}
                color={colors.red}
              />
            ) : null}
            <Menu.Item
              icon='arrow-up-circle'
              title={t('settings.updates-check')}
              onPress={() => navigation.navigate('settings/update')}
            />
          </Menu.Section>
          <Menu.Section>
            <Menu.Item
              icon='terminal'
              title={t('settings.dev-tools')}
              onPress={() => navigation.navigate('settings/devtools')}
            />
          </Menu.Section>
          <Menu.Section>
            <Menu.Item
              icon='lock'
              title={t('settings.security-privacy')}
              onPress={() =>
                navigation.navigate('settings/security-and-privacy')
              }
            />
            <Menu.Item
              icon='send'
              title={t('settings.messages')}
              onPress={() => navigation.navigate('settings/messages-settings')}
            />
            <Menu.Item
              icon='bell'
              title={t('settings.notifications')}
              onPress={() => navigation.navigate('settings/notifications')}
            />
          </Menu.Section>
          <Menu.Section>
            <Menu.Item
              icon='info'
              title={t('settings.about')}
              onPress={() => navigation.navigate('settings/about')}
            />
            <Menu.Item
              icon='activity'
              title={t('settings.news')}
              onPress={() => navigation.navigate('settings/news')}
            />
          </Menu.Section>
          <Menu.Section>
            <Menu.Item
              icon='life-buoy'
              title={t('settings.help')}
              onPress={() => navigation.navigate('settings/help')}
            />
            <Menu.Item
              icon='layers'
              title={t('settings.legal')}
              onPress={() => navigation.navigate('settings/legal')}
            />
          </Menu.Section>
        </Menu>
      )
    }
  )

  render () {
    const { navigation, currentUser, t } = this.props

    return (
      <Screen>
        <UpdateContext.Consumer>
          {({ availableUpdate }) => (
            <List.Menu
              navigation={navigation}
              data={currentUser}
              availableUpdate={availableUpdate}
              t={t}
            />
          )}
        </UpdateContext.Consumer>
      </Screen>
    )
  }
}

export default withCurrentUser(withNamespaces()(List), {
  showOnlyLoaded: true,
  fallback: <Loader />,
})
