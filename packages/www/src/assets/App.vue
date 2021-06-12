<template>
  <div class="container">
    <form id="Filter" @submit.prevent="doFilter">
      <div class="field has-addons mt-4">
        <div class="control is-expanded">
          <input type="search" class="input" name="q" v-model="q" />
        </div>
        <div class="control">
          <button class="button is-info" type="submit">Filter</button>
        </div>
      </div>

      <div
        class="mx-4"
        style="
          display: grid;
          grid-template-columns: minmax(80px, 1.5fr) 1fr 1fr;
          gap: 1em;
          max-width: 800px;
          margin: 0 auto !important;
        "
      >
        <div class="field">
          <div class="control check-list">
            <label class="checkbox">
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
            <label class="checkbox">
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
            <label class="checkbox">
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

        <div class="field" style="justify-self: end; padding-right: 1em">
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

        <div class="field" style="justify-self: end; padding-right: 1em">
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
          <button name="quiz" class="button is-success" type="submit">
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

    <div v-if="isQuiz" class="modal is-active">
      <div class="modal-background"></div>
      <div
        class="modal-card"
        style="min-width: 80vw; max-width: 1200px; height: 100%"
      >
        <header class="modal-card-head">
          <p class="modal-card-title">Quiz</p>
          <button
            class="delete"
            aria-label="close"
            @click="isQuiz = false"
          ></button>
        </header>
        <section class="modal-card-body" style="height: 100%">
          <Quiz
            :session="sessionId"
            @end="
              () => {
                isQuiz = false
                doFilter()
              }
            "
          />
        </section>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { ref, watch, defineComponent, onBeforeMount } from 'vue'
import { makeUseInfiniteScroll } from 'vue-use-infinite-scroll'

import Quiz from './Quiz.vue'
import { api } from './api'

export default defineComponent({
  components: {
    Quiz,
  },
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
    const isQuiz = ref(false)
    const sessionId = ref('')

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
      api
        .post<{
          id: string
        }>('/api/quiz/init', undefined, {
          params: {
            q: q.value,
            state: state.value.join(','),
          },
        })
        .then(({ data }) => {
          sessionId.value = data.id
          isQuiz.value = true
        })
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
      isQuiz,
      sessionId,
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

.check-list {
  display: grid;
  grid-auto-flow: column;
  gap: 1em;

  @media screen and (max-width: 630px) {
    grid-auto-flow: row;
  }
}
</style>
