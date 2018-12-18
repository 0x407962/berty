import React from 'react'
import { View } from 'react-native'
import { StackActions, withNavigation } from 'react-navigation'
import colors from '../../constants/colors'
import Button from './Button'

const ModalScreen = props => {
  const {
    children,
    navigation,
    showDismiss,
    width,
    footer,
    ...otherProps
  } = props

  return <>
    <View style={{
      position: 'absolute',
      top: 0,
      left: 0,
      right: 0,
      bottom: 0,
      zIndex: -1,
    }}>
      <View
        style={{
          backgroundColor: colors.transparentGrey,
          flex: 1,
        }}
      />
    </View>
    <View style={{
      width: width || 320,
      position: 'absolute',
      flex: 1,
    }}>
      <View
        style={[{
          backgroundColor: colors.white,
          borderRadius: 10,
        }]}
        {...otherProps}
      >
        {showDismiss
          ? <View style={{
            flex: 1,
            marginTop: 10,
            marginRight: 10,
            alignSelf: 'flex-end',
            zIndex: 1,
          }}>
            <Button onPress={() => {
              if (props.onDismiss !== undefined) {
                props.onDismiss()
              } else {
                navigation.dispatch(StackActions.pop({
                  n: 1,
                }))
              }
            }} icon={'x'} color={colors.fakeBlack} large />

          </View>
          : null}
        <View style={{
          marginTop: -24,
        }}>
          {children}
        </View>
      </View>
      {footer}
    </View>
  </>
}

export default withNavigation(ModalScreen)
