import ViewShot from 'react-native-view-shot'
import React, { Component } from 'react'
import { View, CameraRoll } from 'react-native'
import { StackActions } from 'react-navigation'

export class ViewExportComponent extends Component {
  async componentDidMount () {
    const resolve = this.props.navigation.getParam('resolve')
    const reject = this.props.navigation.getParam('reject')

    try {
      const uri = await this.refs.viewShot.capture({ format: 'jpg' })

      resolve(uri)
    } catch (e) {
      reject(e)
    }

    this.props.navigation.dispatch(StackActions.pop({
      n: 1,
    }))
  }

  render () {
    const view = this.props.navigation.getParam('view')

    return (
      <View style={{ opacity: 0 }}>
        <ViewShot ref='viewShot'>
          {view}
        </ViewShot>
      </View>
    )
  }
}

export default async ({ view, navigation }) => {
  try {
    const uri = await new Promise((resolve, reject) => {
      navigation.push('virtual/view-export', {
        resolve,
        reject,
        view,
      })
    })
    await CameraRoll.saveToCameraRoll(uri, 'photo')
  } catch (e) {
    throw e
  }
}
