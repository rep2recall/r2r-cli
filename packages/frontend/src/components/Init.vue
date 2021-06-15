<template>
  <component :is="opts.type" v-if="opts.ok" v-bind="opts" />
  <div v-else>Loading...</div>
</template>

<script lang="ts">
import { defineAsyncComponent, defineComponent, ref } from 'vue'

import { initAPI } from '@/assets/api'

export default defineComponent({
  components: {
    App: defineAsyncComponent(() => import('./App.vue')),
    Quiz: defineAsyncComponent(() => import('./Quiz.vue'))
  },
  props: {
    type: {
      type: String,
      required: true
    }
  },
  setup: props => {
    const opts = ref<{
      type: string
      ok?: boolean
    }>({
      type: props.type
    })

    initAPI().then(data => {
      opts.value = Object.assign(opts.value, data)
    })

    return {
      opts
    }
  }
})
</script>
