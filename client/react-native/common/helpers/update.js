import moment from 'moment'
import { DeviceInfos } from '../graphql/queries'
import RNDeviceInfo from 'react-native-device-info'
import { Linking, Platform, PermissionsAndroid } from 'react-native'
import RNFetchBlob from './rn-fetch-blob'
import { showMessage } from 'react-native-flash-message'
import { requestAndroidPermission } from './permissions'

const updateApiSources = {
  'chat.berty.ios': {
    url: 'https://yolo.berty.io/release/ios.json',
    channel: 'dev',
  },
  'chat.berty.house.ios': {
    url: 'https://yolo.berty.io/release/ios-beta.json',
    channel: 'beta',
  },
  'chat.berty.main.debug': {
    url: 'https://yolo.berty.io/release/android.json',
    channel: 'dev',
  },
  'chat.berty.main': {
    url: 'https://yolo.berty.io/release/android.json',
    channel: 'beta',
  },
}

export const getAvailableUpdate = async context => {
  const installedVersion = await getInstalledVersion(context)
  const latestVersion = await getLatestVersion()

  return shouldUpdate(installedVersion, latestVersion) ? latestVersion.installUrl : null
}

export const getInstalledVersion = async context => {
  const bundleId = RNDeviceInfo.getBundleId()

  if (!updateApiSources.hasOwnProperty(bundleId)) {
    return null
  }

  const { channel } = updateApiSources[bundleId]

  const deviceData = await DeviceInfos(context).fetch()
  const [rawVersionInfo] = deviceData.infos.filter(d => d.key === 'versions').map(d => d.value)

  const {
    Core: {
      GitSha: hash,
      GitBranch: branch,
      CommitDate: rawCommitDate,
    },
  } = JSON.parse(rawVersionInfo)

  return {
    channel,
    hash: hash,
    branch: branch,
    buildDate: moment(rawCommitDate, 'YYYY-MM-DD HH:mm:ss ZZ'),
    installUrl: null,
  }
}

export const getLatestVersion = async () => {
  const bundleId = RNDeviceInfo.getBundleId()

  if (!updateApiSources.hasOwnProperty(bundleId)) {
    return null
  }

  const { channel, url } = updateApiSources[bundleId]

  const releases = await fetch(url).then(res => res.json())

  if (!releases.master) {
    return null
  }

  return {
    channel,
    branch: 'master',
    hash: releases.master['git-sha'],
    buildDate: moment(releases.master['stop-time']),
    installUrl: releases.master['manifest-url'],
  }
}

export const installUpdate = async installUrl => {
  if (Platform.OS === 'ios') {
    Linking.openURL(installUrl).catch(e =>
      console.error(e),
    )
  } else if (Platform.OS === 'android') {
    const allowed = await requestAndroidPermission({
      permission: PermissionsAndroid.PERMISSIONS.WRITE_EXTERNAL_STORAGE,
      title: 'Write to external storage',
      message: 'This permission is required to download the application update',
    })

    if (!allowed) {
      showMessage({
        message: 'Unable to download app, not allowed to write file',
        type: 'danger',
        icon: 'danger',
        position: 'top',
      })

      return
    }

    showMessage({
      message: 'Downloading application update',
      type: 'info',
      icon: 'info',
      position: 'top',
    })

    RNFetchBlob
      .config({
        addAndroidDownloads: {
          title: 'berty-update.apk',
          useDownloadManager: true,
          mediaScannable: true,
          notification: true,
          description: 'File downloaded by download manager.',
          path: `${RNFetchBlob.fs.dirs.DownloadDir}/berty-update.apk`,
        },
      })
      .fetch('GET', installUrl)
      .then((res) => {
        RNFetchBlob.android.actionViewIntent(res.path(), 'application/vnd.android.package-archive')
      }).catch(e => {
        showMessage({
          message: String(e),
          type: 'danger',
          icon: 'danger',
          position: 'top',
        })
      })
  }
}

export const shouldUpdate = (installedVersion, latestVersion) => {
  if (!installedVersion || !latestVersion || installedVersion.hash === latestVersion.hash) {
    return false
  }

  return (installedVersion.branch === 'master') ||
    (installedVersion.branch !== 'master' && installedVersion.buildDate.diff(latestVersion.buildDate) < 0)
}
