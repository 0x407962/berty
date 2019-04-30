import EventStream from './EventStream'

export default context => ({
  ...EventStream(context),
  subscribe: ({ updater }) =>
    EventStream(context).subscribe({
      updater:
        updater &&
        (async (store, data) => {
          if (data.EventStream.kind === 302) {
            await context.queries.Conversation.fetch({
              id: data.EventStream.targetAddr,
            })
          }
        }),
    }),
})
