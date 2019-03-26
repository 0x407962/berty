import { NativeModules } from 'react-native'
import { atob } from 'b64-lite'
import React, { PureComponent } from 'react'

import { NavigatorContext } from './NavigatorContext'
import AppNavigator from './AppNavigator'

const { CoreModule } = NativeModules

const getActiveRoute = navigationState => {
  if (!navigationState) {
    return null
  }
  const route = navigationState.routes[navigationState.index]
  // dive into nested navigators
  if (route.routes) {
    return getActiveRoute(route)
  }
  return route
}

const getURIFromRoute = route => {
  // get uri fragment from react-navigation params
  const fragment = Object.keys(route.params || {}).reduce((fragment, key) => {
    const paramType = typeof route.params[key]
    if (
      paramType === 'string' ||
      paramType === 'number' ||
      paramType === 'boolean'
    ) {
      let val = route.params[key]
      try {
        if (key === 'id') {
          val = atob(val)
          val = val.match(/:(.*)$/)
          val = val[1]
        }
      } catch (err) {
        val = route.params[key]
      }
      fragment += fragment.length > 0 ? `,${key}=${val}` : `#${key}=${val}`
    }
    return fragment
  }, '')
  return route.routeName + fragment
}

class Navigator extends PureComponent {
  state = {}

  render () {
    const { screenProps, ...props } = this.props
    return (
      <NavigatorContext.Provider value={this.state}>
        <AppNavigator
          {...props}
          onNavigationStateChange={(prevState, currentState) => {
            const currentRoute = getActiveRoute(currentState)
            const prevRoute = getActiveRoute(prevState)
            if (prevRoute !== currentRoute) {
              CoreModule.setCurrentRoute(getURIFromRoute(currentRoute))
              this.setState(currentRoute)
            }
          }}
        />
      </NavigatorContext.Provider>
    )
  }
}

export default Navigator
