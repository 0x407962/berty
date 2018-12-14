import { Platform, TextInput as RNTextInput } from 'react-native'
import React, { PureComponent } from 'react'

import { Pagination, RelayContext } from '../../../relay'
import { Text, Flex, Screen, Header, Icon } from '../../Library'
import { colors } from '../../../constants'
import { fragments } from '../../../graphql'
import { merge } from '../../../helpers'
import { parseEmbedded } from '../../../helpers/json'
import { shadow } from '../../../styles'
import { conversation as utils } from '../../../utils'

class Message extends React.PureComponent {
  static contextType = RelayContext

  messageSeen = async () => {
    await this.props.screenProps.context.mutations.eventSeen({
      id: this.props.data.id,
    })
  }

  render () {
    const conversation = this.props.navigation.getParam('conversation')
    const contactId = this.props.data.senderId
    const isMyself =
      conversation.members.find(m => m.contactId === contactId).contact
        .status === 42

    const { data } = this.props

    // TODO: implement message seen
    // if (new Date(this.props.data.seenAt).getTime() <= 0) {
    //   this.messageSeen()
    // }
    return (
      <Flex.Rows
        align={isMyself ? 'end' : 'start'}
        style={{ marginHorizontal: 10, marginVertical: 10 }}
      >
        <Text
          padding={{
            vertical: 6,
            horizontal: 10,
          }}
          multiline
          left={!isMyself}
          right={isMyself}
          background={colors.blue}
          color={colors.white}
          rounded={14.5}
          margin={{
            bottom: 4,
            [isMyself ? 'left' : 'right']: 42,
          }}
        >
          {parseEmbedded(data.attributes).message.text}
        </Text>
        <Text
          left={!isMyself}
          right={isMyself}
          tiny
          color={colors.subtleGrey}
          margin={{
            top: 6,
            bottom: 6,
            [isMyself ? 'left' : 'right']: 42,
          }}
        >
          {new Date(data.createdAt).toTimeString()}{' '}
          {isMyself ? (
            <Icon
              name={
                new Date(data.ackedAt).getTime() > 0 ? 'check-circle' : 'circle'
              }
            />
          ) : null}{' '}
          <Icon
            name={new Date(data.seenAt).getTime() > 0 ? 'eye' : 'eye-off'}
          />{' '}
          {/* TODO: used for debugging, remove me */}
        </Text>
      </Flex.Rows>
    )
  }
}

const MessageContainer = fragments.Event(Message)

class TextInput extends PureComponent {
  state = {
    height: 16,
    value: '',
  }

  onContentSizeChange = ({
    nativeEvent: {
      contentSize: { height },
    },
  }) => this.setState({ height: height > 80 ? 80 : height })

  render () {
    const { height } = this.state
    const { value } = this.props
    return (
      <RNTextInput
        style={[
          {
            flex: 1,
            padding: 0,
            marginVertical: 8,
            marginHorizontal: 0,
            height: height,
          },
          Platform.OS === 'web' ? { paddingLeft: 16 } : {},
        ]}
        onContentSizeChange={this.onContentSizeChange}
        autoFocus
        placeholder='Write a secure message...'
        onChangeText={this.props.onChangeText}
        value={value}
        multiline
      />
    )
  }
}

class Input extends PureComponent {
  static contextType = RelayContext

  state = {
    input: '',
  }

  async componentDidMount () {
    const conversation = this.props.navigation.getParam('conversation')
    await this.props.screenProps.context.mutations.conversationRead({
      id: conversation.id,
    })
  }

  async componentWillUnmount () {
    const conversation = this.props.navigation.getParam('conversation')
    await this.props.screenProps.context.mutations.conversationRead({
      id: conversation.id,
    })
  }

  onSubmit = () => {
    const { input } = this.state
    this.setState({ input: '' }, async () => {
      try {
        const conversation = this.props.navigation.getParam('conversation')
        await this.props.screenProps.context.mutations.conversationAddMessage({
          conversation: {
            id: conversation.id,
          },
          message: {
            text: input,
          },
        })
      } catch (err) {
        console.error(err)
      }
    })
  }
  onChangeText = value => this.setState({ input: value })

  render () {
    return (
      <Flex.Cols
        size={0}
        justify='center'
        align='center'
        style={
          Platform.OS === 'web'
            ? [{ position: 'absolute', bottom: 0, left: 0, right: 0 }, shadow]
            : [shadow]
        }
      >
        <Flex.Cols
          style={{
            backgroundColor: colors.grey8,
            marginLeft: 8,
            marginRight: 3,
            borderRadius: 16,
            marginVertical: 8,
          }}
        >
          <Text
            left
            top
            size={0}
            icon='edit-2'
            padding={{ right: 5, left: 8, vertical: 7 }}
          />
          <TextInput
            onChangeText={this.onChangeText}
            value={this.state.input}
          />
        </Flex.Cols>
        <Text
          right
          size={0}
          middle
          margin={{ right: 8, ...(Platform.OS === 'web' ? { left: 12 } : {}) }}
          padding
          large
          icon='send'
          color={colors.grey5}
          onPress={this.onSubmit}
        />
      </Flex.Cols>
    )
  }
}

export default class Detail extends PureComponent {
  static navigationOptions = ({ navigation }) => ({
    header: (
      <Header
        navigation={navigation}
        title={utils.getTitle(navigation.getParam('conversation'))}
        backBtn
        rightBtnIcon='settings'
        onPressRightBtn={() =>
          navigation.push('chats/settings', {
            conversation: navigation.getParam('conversation'),
          })
        }
      />
    ),
  })

  render () {
    const conversation = this.props.navigation.getParam('conversation')
    const {
      navigation,
      screenProps: {
        context: { queries, subscriptions },
      },
    } = this.props

    return (
      <Screen style={{ backgroundColor: colors.white, paddingTop: 0 }}>
        <Flex.Rows>
          <Pagination
            style={[
              { flex: 1 },
              Platform.OS === 'web' ? { paddingTop: 48 } : {},
            ]}
            direction='forward'
            query={queries.EventList.graphql}
            variables={merge([
              queries.EventList.defaultVariables,
              {
                filter: {
                  kind: 302,
                  conversationId: conversation.id,
                },
              },
            ])}
            subscriptions={[subscriptions.newMessage]}
            fragment={fragments.EventList}
            alias='EventList'
            renderItem={props => (
              <MessageContainer
                {...props}
                navigation={navigation}
                screenProps={this.props.screenProps}
              />
            )}
            inverted
          />
          <Input
            navigation={this.props.navigation}
            screenProps={this.props.screenProps}
          />
        </Flex.Rows>
      </Screen>
    )
  }
}
