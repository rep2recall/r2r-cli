<template>
  <div class="container">
    <form id="Filter" class="field has-addons mt-4" @submit.prevent="doFilter">
      <div class="control is-expanded">
        <input type="search" class="input" name="q" v-model="q" />
      </div>
      <div class="control">
        <button class="button is-primary" type="submit">Filter</button>
      </div>
    </form>

    <div id="Stat" class="card mt-4">
      <header class="card-header">
        <h2 class="card-header-title">Quiz statistics</h2>
      </header>

      <div class="card-content">
        <div class="columns">
          <div class="column is-4">
            <label for="new">New:</label>
            <span name="new">
              {{ stat.new }}
            </span>
          </div>
          <div class="column is-4">
            <label for="due">Due:</label>
            <span name="due">
              {{ stat.due }}
            </span>
          </div>
          <div class="column is-4">
            <label for="leech">Leech:</label>
            <span name="leech">
              {{ stat.leech }}
            </span>
          </div>
        </div>

        <form class="field" @submit.prevent="doQuiz">
          <button name="quiz" class="button is-primary" type="button">
            Quiz
          </button>
        </form>
      </div>
    </div>

    <div id="Leech" class="card mt-4" v-show="leechItems.length">
      <header class="card-header">
        <h2 class="card-header-title">Leeches</h2>
        <div class="card-header-icon">
          <button
            class="delete"
            type="button"
            @click="isLeechOpen = !isLeechOpen"
          ></button>
        </div>
      </header>
      <div
        v-show="isLeechOpen"
        class="card-content"
        style="max-height: 400px; overflow: auto"
      >
        <div class="columns" style="flex-wrap: wrap">
          <div class="column is-4" v-for="id in leechItems" :key="id">
            <iframe
              :src="'/item.html?id=' + id"
              style="height: 200px; width: 100%; border: 1px solid lightgray"
              sandbox="allow-scripts"
            ></iframe>
          </div>

          <div ref="scrollTrigger"></div>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { ref, watch, defineComponent } from 'vue'
import { makeUseInfiniteScroll } from 'vue-use-infinite-scroll'

export default defineComponent({
  setup() {
    const q = ref(
      (() => {
        const u = new URL(location.href)
        return u.searchParams.get('q') || ''
      })()
    )
    const stat = ref({
      new: 0,
      due: 0,
      leech: 0,
    })
    const leechItems = ref([])
    const isLeechOpen = ref(false)

    const useInfiniteScroll = makeUseInfiniteScroll({})
    const scrollTrigger = ref<HTMLElement>(null)
    const scrollRef = useInfiniteScroll(scrollTrigger)

    watch(
      scrollRef,
      (p) => {
        leechItems.value = [
          ...leechItems.value,
          ...Array(6)
            .fill(null)
            .map(() => new Date().toISOString()),
        ]
      },
      { immediate: true }
    )

    const doFilter = () => {
      console.log('Filtering', q.value)
    }

    const doQuiz = () => {
      console.log('Quizzing', q.value)
    }

    return {
      q,
      doFilter,
      stat,
      doQuiz,
      leechItems,
      scrollTrigger,
      isLeechOpen,
    }
  },
})
</script>

<style lang="scss">
#Stat .card-content {
  display: grid;
  justify-content: center;
  grid-template-columns: 1fr auto;

  .columns {
    margin: 0;
  }
}
</style>
