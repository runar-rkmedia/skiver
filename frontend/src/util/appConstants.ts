const appConstants = {
  baseUrl: `${window.location.protocol}//${window.location.host}${window.location.pathname === '/skiver/' ? '/skiver' : ''}`
}


export const appUrl = (path: string) => appConstants.baseUrl + path
export const apiUrl = (path: string) => appConstants.baseUrl  +"/api" + path


export default appConstants
