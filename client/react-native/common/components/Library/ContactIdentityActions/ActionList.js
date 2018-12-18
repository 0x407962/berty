import React, { PureComponent } from 'react'
import { withNavigation } from 'react-navigation'
import ActionButton from './ActionButton'
import colors from '../../../constants/colors'
import { Flex } from '../index'
import { showMessage } from 'react-native-flash-message'

class ActionList extends PureComponent {
  render = () => {
    let { children, inModal } = this.props
    const count = React.Children.count(children)

    if (count > 4) {
      children = React.Children.toArray(children).slice(0, 4)
    }

    const large = count < 3

    return <Flex.Cols>
      {React.Children.map(children, action => <action.type {...action.props} large={large}
        dismissOnSuccess={inModal && action.props.dismissOnSuccess} />)}
    </Flex.Cols>
  }
}

class Action extends PureComponent {
  constructor (props) {
    super(props)

    this.state = {
      loading: false,
      success: false,
    }

    this._mounted = false
  }

  componentWillUnmount () {
    this._mounted = false
  }

  componentDidMount () {
    this._mounted = true
  }

  render = () => {
    const { large, icon, title, color = colors.blue, action, dismissOnSuccess, navigation, successMessage, successType } = this.props
    const ButtonClass = large ? ActionButton.Large : ActionButton

    return <ButtonClass icon={icon} title={title} color={color}
      onPress={!this.state.loading && (async () => {
        await new Promise(resolve => this.setState({ loading: true }, resolve))

        try {
          await action()

          if (successMessage) {
            showMessage({
              message: successMessage,
              type: successType || 'info',
              icon: successType || 'info',
              position: 'top',
            })
          }

          if (dismissOnSuccess) {
            navigation.goBack(null)
          }

          if (this._mounted) {
            await new Promise(resolve => this.setState({ loading: false }, resolve))
          }
        } catch (e) {
          showMessage({
            message: String(e),
            type: 'danger',
            icon: 'danger',
            position: 'top',
          })

          await new Promise(resolve => this.setState({ loading: false }, resolve))
        }
      })} />
  }
}

export default ActionList
ActionList.Action = withNavigation(Action)
