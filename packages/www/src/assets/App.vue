<template>
  <div class="container">
    <form id="Filter" @submit.prevent="doFilter">
      <div class="field has-addons mt-4">
        <div class="control is-expanded">
          <input type="search" class="input" name="q" v-model="q" />
        </div>
        <div class="control">
          <button class="button is-primary" type="submit">Filter</button>
        </div>
      </div>

      <div
        class="mx-4"
        style="display: flex; flex-direction: row; justify-content: center"
      >
        <div class="field">
          <div class="control">
            <label class="radio">
              <input
                type="checkbox"
                name="learning"
                :checked="state.includes('new')"
                @change="
                  (ev) =>
                    (state = ev.target.checked
                      ? [...state, 'new']
                      : state.filter((r) => r !== 'new'))
                "
              />
              New
            </label>
            <label class="radio">
              <input
                type="checkbox"
                name="learning"
                :checked="state.includes('learning')"
                @change="
                  (ev) =>
                    (state = ev.target.checked
                      ? [...state, 'learning']
                      : state.filter((r) => r !== 'learning'))
                "
              />
              Learning
            </label>
            <label class="radio">
              <input
                type="checkbox"
                name="learning"
                :checked="state.includes('graduated')"
                @change="
                  (ev) =>
                    (state = ev.target.checked
                      ? [...state, 'graduated']
                      : state.filter((r) => r !== 'graduated'))
                "
              />
              Graduated
            </label>
          </div>
        </div>

        <div style="width: 5em"></div>

        <div class="field">
          <label class="checkbox">
            <input
              type="checkbox"
              :checked="!state.includes('due')"
              @change="
                (ev) =>
                  (state = !ev.target.checked
                    ? [...state, 'due']
                    : state.filter((r) => r !== 'due'))
              "
            />
            Include undue
          </label>
        </div>

        <div style="width: 5em"></div>

        <div class="field">
          <label class="checkbox">
            <input
              type="checkbox"
              :checked="state.includes('leech')"
              @change="
                (ev) =>
                  (state = ev.target.checked
                    ? [...state, 'leech']
                    : state.filter((r) => r !== 'leech'))
              "
            />
            Include leeches
          </label>
        </div>
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
            <div data-stat name="new">
              {{ stat.new.toLocaleString() }}
            </div>
          </div>
          <div class="column is-4">
            <label for="due">Due:</label>
            <div data-stat name="due">
              {{ stat.due.toLocaleString() }}
            </div>
          </div>
          <div class="column is-4">
            <label for="leech">Leech:</label>
            <div data-stat name="leech">
              {{ stat.leech.toLocaleString() }}
            </div>
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
      <header
        class="card-header"
        @click="isLeechOpen = !isLeechOpen"
        style="cursor: pointer"
      >
        <h2 class="card-header-title">Leeches</h2>
        <div class="card-header-icon">
          <button class="delete" type="button"></button>
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
              :src="`/card.html?side=front&id=${id}&secret=${secret}`"
              style="height: 200px; width: 100%; border: 1px solid lightgray"
              sandbox="allow-scripts allow-same-origin allow-forms"
            ></iframe>
          </div>

          <div ref="scrollTrigger"></div>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { ref, watch, defineComponent, onBeforeMount } from 'vue'
import { makeUseInfiniteScroll } from 'vue-use-infinite-scroll'

import { api } from './api'

export default defineComponent({
  setup() {
    const q = ref(
      (() => {
        const u = new URL(location.href)
        return u.searchParams.get('q') || ''
      })()
    )
    const state = ref(['new', 'learning', 'due'])

    const stat = ref({
      new: 0,
      due: 0,
      leech: 0,
    })
    const leechItems = ref([] as string[])
    const isLeechOpen = ref(false)

    const useInfiniteScroll = makeUseInfiniteScroll({})
    const scrollTrigger = ref(null as any)
    const scrollRef = useInfiniteScroll(scrollTrigger)

    watch(
      scrollRef,
      (page) => {
        const limit = 6

        api
          .get<{
            result: string[]
          }>('/api/quiz/leech', {
            params: {
              page,
              limit,
              q: q.value,
            },
          })
          .then(({ data }) => {
            leechItems.value = [...leechItems.value, ...data.result]
          })
      },
      { immediate: true }
    )

    const doFilter = () => {
      leechItems.value = []
      scrollRef.value = Number(!!scrollRef.value)

      api
        .get<{
          new: number
          due: number
          leech: number
        }>('/api/quiz/stat', {
          params: {
            q: q.value,
            state: state.value.join(','),
          },
        })
        .then(({ data }) => {
          stat.value = data
        })
    }

    const doQuiz = () => {
      console.log('Quizzing', q.value)
    }

    onBeforeMount(() => {
      doFilter()
    })

    watch(state, () => {
      doFilter()
    })

    return {
      q,
      state,
      doFilter,
      stat,
      doQuiz,
      leechItems,
      scrollTrigger,
      isLeechOpen,
      secret: new URL(location.href).searchParams.get('secret'),
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

  [data-stat] {
    display: inline-block;
    width: 1.5em;
    text-align: right;

    label + & {
      margin-left: 0.5em;
    }
  }
}
</style>
