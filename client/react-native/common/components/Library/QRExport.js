import React from 'react'
import { Text } from 'react-native'
import ContactIdentity from './ContactIdentity'

const QRCodeExport = ({ data }) => <>
  <Text>{data.displayName} is on Berty</Text>
  <ContactIdentity.QrCode data={data} />
</>

export default QRCodeExport
