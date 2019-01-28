import { View } from 'react-native'
import React, { PureComponent } from 'react'
import { Button, Flex, Text, SearchBar } from '.'
import { colors } from '../../constants'
import { padding, borderBottom, paddingBottom } from '../../styles'
import { isRTL } from '../../i18n'
import RelayContext from '../../relay/RelayContext'
import Icon from './Icon'
import { promiseWithTimeout } from 'react-relay-network-modern/es/middlewares/retry'

const [defaultTextColor, defaultBackColor] = [colors.black, colors.white]

const HeaderButton = ({ icon, color, style, ...otherProps }) => {
  return <Button icon={icon} large color={color} {...otherProps} />
}

class StateBadge extends PureComponent {
  constructor (props) {
    super(props)
    this.state = {
      watchTime: 10000,
      listenAddrs: [],
      listenInterfaceAddrs: [],
      timeouted: false,
      requestTimeout: 3000,
      color: colors.black,
      listenAddrTimer: null,
      InterfaceAddrTimer: null,
    }
  }

  componentDidMount () {
    this.fetchListenAddrs()
    this.fetchListenInterfaceAddrs()
  }

  componentWillUnmount () {
    const { listenAddrTimer, InterfaceAddrTimer } = this.state

    if (listenAddrTimer !== null) {
      clearTimeout(listenAddrTimer)
    }

    if (InterfaceAddrTimer !== null) {
      clearTimeout(InterfaceAddrTimer)
    }
  }

  timeoutPromise = () => {
    return new Promise((resolve, reject) => {
      this.setState({ timeouted: true }, this.setColor)
      reject(new Error('Request timed out'))
    })
  }

  fetchListenAddrs = () => {
    const { context } = this.props
    const { watchTime, requestTimeout } = this.state

    promiseWithTimeout(context.queries.GetListenAddrs.fetch(), requestTimeout, this.timeoutPromise).then(e => {
      const timer = setTimeout(this.fetchListenAddrs, watchTime)
      this.setState({ listenAddrs: e.addrs, timeouted: false, listenAddrTimer: timer }, this.setColor)
    }).catch(err => {
      const timer = setTimeout(this.fetchListenAddrs, watchTime)
      this.setState({ listenAddrTimer: timer })
      console.log('err Listen address', err)
    })
  }

  fetchListenInterfaceAddrs = () => {
    const { context } = this.props
    const { watchTime, requestTimeout } = this.state

    promiseWithTimeout(context.queries.GetListenInterfaceAddrs.fetch(), requestTimeout, this.timeoutPromise).then(e => {
      const timer = setTimeout(this.fetchListenInterfaceAddrs, watchTime)
      this.setState({ listenInterfaceAddrs: e.addrs, timeouted: false, InterfaceAddrTimer: timer }, this.setColor)
    }).catch(err => {
      const timer = setTimeout(this.fetchListenInterfaceAddrs, watchTime)
      this.setState({ InterfaceAddrTimer: timer })
      console.log('err Listen address', err)
    })
  }

  setColor = () => {
    const { listenAddrs, listenInterfaceAddrs, timeouted } = this.state
    let color = colors.black

    if (listenAddrs.length > 0 && listenInterfaceAddrs.length > 0) {
      color = colors.yellow
      listenInterfaceAddrs.forEach((v, i, arr) => {
        try {
          const splited = v.split('/')
          if (splited[1] === 'ip4' && splited[2] !== '127.0.0.1') {
            color = colors.green
          }
        } catch (e) {
          // Silence error since /p2p-circuit isn't valid
          // console.log(e)
        }
      })
    }

    if (timeouted) {
      color = colors.red
    }
    this.setState({ color })
  }

  render () {
    const { color } = this.state

    return (<Icon style={{ color }} name={'material-checkbox-blank-circle'} />)
  }
}

export default class Header extends PureComponent {
  render () {
    const {
      navigation,
      title,
      titleIcon,
      backBtn,
      rightBtn,
      rightBtnIcon,
      onPressRightBtn,
      searchBar,
      searchHandler,
    } = this.props

    const colorText =
      this.props.colorText == null ? defaultTextColor : this.props.colorText
    const colorBack =
      this.props.colorBack == null ? defaultBackColor : this.props.colorBack
    const colorBtnLeft =
      this.props.colorBtnLeft == null
        ? defaultTextColor
        : this.props.colorBtnLeft
    const colorBtnRight =
      this.props.colorBtnRight == null
        ? defaultTextColor
        : this.props.colorBtnRight

    let searchBarComponent = null
    if (searchBar === true) {
      searchBarComponent = (
        <SearchBar onChangeText={text => searchHandler(text)} />
      )
    } else if (searchBar !== undefined && searchBar !== false) {
      searchBarComponent = searchBar
    }

    return (
      <View
        style={[
          {
            backgroundColor: colorBack,
            height: searchBar ? 100 : 60,
          },
          borderBottom,
          padding,
        ]}
      >
        <Flex.Rows>
          <Flex.Cols
            size={1}
            justify='between'
            align='center'
            style={[searchBar ? paddingBottom : {}]}
          >
            {backBtn && (
              <HeaderButton
                icon='arrow-left'
                color={colorBtnLeft}
                onPress={() => {
                  if (typeof backBtn === 'function') {
                    backBtn()
                  }
                  navigation.goBack(null)
                }}
                flip={isRTL()}
                justify='start'
                middle
              />
            )}
            <Text
              icon={titleIcon}
              left
              large
              color={colorText}
              justify={backBtn ? 'center' : 'start'}
              middle
              size={5}
            >
              {navigation.state.routeName === 'chats/list'
                ? <RelayContext.Consumer>
                  {context => <StateBadge context={context} />}
                </RelayContext.Consumer>
                : null
              }
              {title}
            </Text>
            {rightBtn ? <View>{rightBtn}</View> : null}
            {!rightBtn &&
              rightBtnIcon !== null && (
              <HeaderButton
                icon={rightBtnIcon}
                color={colorBtnRight}
                onPress={onPressRightBtn}
                justify='end'
                middle
              />
            )}
          </Flex.Cols>

          {searchBarComponent}
        </Flex.Rows>
      </View>
    )
  }
}

Header.HeaderButton = HeaderButton
