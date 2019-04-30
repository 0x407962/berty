import { fetchQuery, graphql } from 'react-relay'

import { event } from '@berty/common/utils'
import { merge } from '@berty/common/helpers'

const query = graphql`
  query EventListQuery(
    $filter: BertyEntityEventInput
    $orderBy: String!
    $orderDesc: Bool!
    $count: Int32
    $cursor: String
    $onlyWithoutAckedAt: Enum
    $onlyWithoutSeenAt: Enum
  ) {
    ...EventList
      @arguments(
        filter: $filter
        orderBy: $orderBy
        orderDesc: $orderDesc
        count: $count
        cursor: $cursor
        onlyWithoutAckedAt: $onlyWithoutAckedAt
        onlyWithoutSeenAt: $onlyWithoutSeenAt
      )
  }
`

export const defaultVariables = {
  filter: event.default,
  orderBy: 'created_at',
  orderDesc: true,
  count: 20,
  cursor: '',
  onlyWithoutAckedAt: 0,
  onlyWithoutSeenAt: 0,
}

export default context => ({
  graphql: query,
  defaultVariables,
  fetch: async variables =>
    fetchQuery(
      context.environment,
      query,
      merge([defaultVariables, variables])
    ),
})
