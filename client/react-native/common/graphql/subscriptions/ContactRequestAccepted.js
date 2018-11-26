import EventStream from './EventStream'

export default context => ({
  ...EventStream(context),
  subscribe: ({ updater }) =>
    EventStream(context).subscribe({
      updater:
        updater &&
        ((store, data) => {
          if (data.EventStream.kind === 202) {
            return updater()
          }
        }),
    }),
})
