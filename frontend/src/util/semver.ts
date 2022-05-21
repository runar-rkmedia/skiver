
// Very naive regex for semver.
const semverReges = /^\d+\.\d+\.\d+-?/

const isSemver = (s: string) => {
  return semverReges.test(s)
}

export default isSemver
