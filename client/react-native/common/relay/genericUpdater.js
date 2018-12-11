import { atob } from 'b64-lite'
import Case from 'case'

import { ConnectionHandler } from 'relay-runtime'

import { FragmentHelper } from './fragment-helper'
import { merge } from '../helpers'

const deepFilterEqual = (a, b) => {
  if (!a) {
    return true
  }
  if (typeof a !== typeof b) {
    return false
  }
  switch (typeof a) {
    case 'object':
      if (Array.isArray(a)) {
        return a.every(av => b.some(bv => deepFilterEqual(av, bv)))
      }
      return Object.keys(a).every(k => deepFilterEqual(a[k], b[k]))
    default:
      return a === b
  }
}

//
//  This function will create a generic relay updater for pagination connection
//  based on fragment information
//  The filter "args" need to have same field that the connection have in arguments
//
export default (fragment, alias, args) => {
  return (store, data, deletion) => {
    const helper = new FragmentHelper(fragment)
    const connectionHelper = helper.getConnection(alias)
    const root = store.getRoot()
    console.log({ fragment, alias, args })
    console.log({ store, data, deletion })
    const connection = ConnectionHandler.getConnection(
      root,
      helper.getConnection(alias).key,
      args
    )
    if (
      deletion ||
      deepFilterEqual(args, merge([args, { filter: data }])) === false
    ) {
      // delete
      ConnectionHandler.deleteNode(connection, data.id)
      return
    }

    // get all edges
    const node =
      store.get(data.id) ||
      store.create(data.id, connectionHelper.getEdgeNodeType())
    console.log('node', node)
    const edges = connection.getLinkedRecords('edges')
    const field = Case.camel(args.orderBy || args.sortBy || 'id')
    console.log('btoa', node.getValue('id'))
    const cursor =
      field === 'id'
        ? atob(node.getValue('id')).split(/:(.+)/)[1]
        : node.getValue(field)
    console.log('cursor', cursor)
    if (
      edges.length > 0 &&
      edges.some(
        e =>
          console.log('edges: cursor: ', e.getValue('cursor')) ||
          e.getValue('cursor') === cursor
      )
    ) {
      // update
      return
    }

    console.log('add connection', { connection })
    // add
    const edge = ConnectionHandler.createEdge(
      store,
      connection,
      node,
      connectionHelper.getEdgeType()
    )
    edge.setValue(cursor, 'cursor')

    console.log(store)
    if (connectionHelper.direction === 'forward' && args.orderDesc === false) {
      console.log('forward')
      ConnectionHandler.insertEdgeAfter(connection, edge, cursor)
      return
    }
    console.log('backward')
    ConnectionHandler.insertEdgeBefore(connection, edge, cursor)
  }
}
