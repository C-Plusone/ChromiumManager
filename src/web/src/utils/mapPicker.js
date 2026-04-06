import { reactive } from 'vue'

export const state = reactive({
  visible: false,
  location: '',
  readonly: false
})

let resolvePromise = null

export const resolve = (loc) => {
  state.visible = false
  if (resolvePromise) {
    resolvePromise(loc)
    resolvePromise = null
  }
}

export const openMapPicker = (loc = '') => {
  state.location = loc
  state.readonly = false
  state.visible = true

  return new Promise((r) => {
    resolvePromise = r
  })
}

export const viewMapLocation = (loc) => {
  state.location = loc
  state.readonly = true
  state.visible = true
}
