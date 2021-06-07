<template>
  <component :is="opts.type" v-if="opts.ok" :options="opts" />
  <div v-else>Loading...</div>
</template>

<script lang="ts">
import { defineAsyncComponent, defineComponent, ref } from 'vue'
import { initAPI } from './assets/api'

export default defineComponent({
  components: {
    App: defineAsyncComponent(() => import('./App.vue')),
  },
  props: ['type'],
  setup: ({ type }: { type: string }) => {
    const opts = ref<{
      type: string
      ok?: boolean
    }>({
      type,
    })

    initAPI().then((data) => {
      opts.value = Object.assign(opts.value, data)
    })

    return {
      opts,
    }
  },
})
</script>
