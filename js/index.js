/**
 * @format
 */

import 'node-libs-react-native/globals'
import 'react-native-gesture-handler'
import { AppRegistry } from 'react-native'

import App from '@berty-tech/messenger-app/App'
import { name as appName } from '@berty-tech/messenger-app/app.json'

if (typeof Buffer === 'undefined') {
	global.Buffer = require('buffer').Buffer
}

AppRegistry.registerComponent(appName, () => App)
