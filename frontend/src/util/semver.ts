
// Very naive regex for semver.
const semverReges = /^\d(?:\.0)?(?:\.0)?-?/

const isPartialSemver = (s: string) => {
  return semverReges.test(s)
}

export default isPartialSemver
