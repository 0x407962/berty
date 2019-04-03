import { graphql, createFragmentContainer } from 'react-relay'

export default component =>
  createFragmentContainer(
    component,
    graphql`
      fragment Conversation on BertyEntityConversation {
        id
        createdAt
        updatedAt
        readAt
        title
        topic
        infos
        kind
        members {
          id
          createdAt
          updatedAt
          readAt
          status
          contact {
            id
            createdAt
            updatedAt
            sigchain
            status
            devices {
              id
              createdAt
              updatedAt
              name
              status
              apiVersion
              contactId
            }
            displayName
            displayStatus
            overrideDisplayName
            overrideDisplayStatus
          }
          conversationId
          contactId
        }
      }
    `
  )
