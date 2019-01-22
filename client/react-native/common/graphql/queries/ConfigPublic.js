import { fetchQuery, graphql } from 'react-relay'
import { merge } from '../../helpers'

const query = graphql`
  query ConfigPublicQuery($t: Bool!) {
    ConfigPublic(T: $t) {
      id
      createdAt
      updatedAt
      pushRelayPubkeyApns,
      pushRelayPubkeyFcm,
    }
  }
`

const defaultVariables = {
  t: true,
}

export default context => ({
  graphql: query,
  defaultVariables,
  fetch: async variables =>
    (await fetchQuery(
      context.environment,
      query,
      merge([defaultVariables, variables])
    )).ConfigPublic,
})
