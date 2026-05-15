export const mobileQaViewports = [
  { width: 390, height: 844, label: 'iPhone 12/13 portrait' },
  { width: 375, height: 812, label: 'iPhone X/11 Pro portrait' },
  { width: 360, height: 740, label: 'Small Android portrait' },
]

export const mobileQaChecks = [
  'top-nav-clearance',
  'bottom-nav-clearance',
  'no-horizontal-overflow',
  'primary-content-visible',
]

export const mobileQaPages = [
  {
    route: '/',
    componentName: 'HomePage',
    sourcePath: 'src/pages/HomePage.tsx',
    pageShellPattern: /className="home-page page"/,
  },
  {
    route: '/profile',
    componentName: 'ProfilePage',
    sourcePath: 'src/pages/ProfilePage.tsx',
    pageShellPattern: /className="profile-page container page"/,
  },
  {
    route: '/history',
    componentName: 'HistoryPage',
    sourcePath: 'src/pages/HistoryPage.tsx',
    pageShellPattern: /className="history-page page"/,
  },
  {
    route: '/compatibility',
    componentName: 'CompatibilityPage',
    sourcePath: 'src/pages/CompatibilityPage.tsx',
    pageShellPattern: /className="page compatibility-page"/,
  },
  {
    route: '/compatibility/history',
    componentName: 'CompatibilityHistoryPage',
    sourcePath: 'src/pages/CompatibilityHistoryPage.tsx',
    pageShellPattern: /className="compatibility-history-page page"/,
  },
  {
    route: '/history/:id',
    componentName: 'ResultPage',
    sourcePath: 'src/pages/ResultPage.tsx',
    pageShellPattern: /className="result-page page/,
  },
  {
    route: '/compatibility/:id',
    componentName: 'CompatibilityResultPage',
    sourcePath: 'src/pages/CompatibilityResultPage.tsx',
    pageShellPattern: /className="page compatibility-result-page"/,
  },
]
