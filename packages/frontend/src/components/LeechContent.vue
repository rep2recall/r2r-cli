<template>
  <div id="LeechContent" class="card-content">
    <div class="columns">
      <div class="column is-4" v-for="id in leechItems" :key="id">
        <iframe
          class="leech-iframe"
          :src="`/card?side=front&id=${id}&secret=${secret}`"
        ></iframe>
      </div>

      <div ref="scrollTrigger"></div>
    </div>
  </div>
</template>

<script lang="ts">
import { api } from '@/assets/api'
import { defineComponent, ref, watch } from 'vue'
import { makeUseInfiniteScroll } from 'vue-use-infinite-scroll'

export default defineComponent({
  props: {
    q: {
      type: String,
      required: true
    }
  },
  setup(props) {
    const leechItems = ref([] as string[])

    const useInfiniteScroll = makeUseInfiniteScroll({})
    const scrollTrigger = ref(null as any)
    const scrollRef = useInfiniteScroll(scrollTrigger)

    watch(
      scrollRef,
      page => {
        const limit = 6

        api
          .get<{
            result: string[]
          }>('/api/quiz/leech', {
            params: {
              page,
              limit,
              q: props.q
            }
          })
          .then(({ data }) => {
            leechItems.value = [...leechItems.value, ...data.result]
          })
      },
      { immediate: true }
    )

    return {
      leechItems,
      scrollTrigger,
      secret: new URL(location.href).searchParams.get('secret')
    }
  }
})
</script>

<style lang="scss" scoped>
#LeechContent {
  max-height: 400px;
  overflow: auto;

  > .columns {
    flex-wrap: wrap;

    .leech-iframe {
      height: 200px;
      width: 100%;
      border: 1px solid lightgray;
    }
  }
}
</style>
