const { history } = require('./data/demo')

App({
  globalData: {
    previewHistoryNext: false,
    previewHistory: history.map(item => ({ ...item })),
    previewQuery: null,
    previewDayProfileEnabled: false,
  },
})
