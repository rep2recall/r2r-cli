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

      <div id="State">
        <div class="field">
          <div class="control check-list">
            <label class="checkbox">
              <input
                type="checkbox"
                name="learning"
                :checked="state.includes('new')"
                @change="
                  ev =>
                    (state = ev.target.checked
                      ? [...state, 'new']
                      : state.filter(r => r !== 'new'))
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
                  ev =>
                    (state = ev.target.checked
                      ? [...state, 'learning']
                      : state.filter(r => r !== 'learning'))
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
                  ev =>
                    (state = ev.target.checked
                      ? [...state, 'graduated']
                      : state.filter(r => r !== 'graduated'))
                "
              />
              Graduated
            </label>
          </div>
        </div>

        <div class="field">
          <label class="checkbox">
            <input
              type="checkbox"
              :checked="!state.includes('due')"
              @change="
                ev =>
                  (state = !ev.target.checked
                    ? [...state, 'due']
                    : state.filter(r => r !== 'due'))
              "
            />
            Include undue
          </label>
        </div>

        <div class="field">
          <label class="checkbox">
            <input
              type="checkbox"
              :checked="state.includes('leech')"
              @change="
                ev =>
                  (state = ev.target.checked
                    ? [...state, 'leech']
                    : state.filter(r => r !== 'leech'))
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
          <button
            :disabled="!(stat.next || stat.due)"
            name="quiz"
            class="button is-success"
            type="submit"
          >
            Quiz
          </button>
        </form>
      </div>
    </div>

    <div id="Leech" class="card mt-4" v-show="!!stat.leech">
      <header class="card-header clickable" @click="isLeechOpen = !isLeechOpen">
        <h2 class="card-header-title">Leeches</h2>
        <div class="card-header-icon">
          <button class="delete" type="button"></button>
        </div>
      </header>

      <leech-content v-if="isLeechOpen" :q="q" />
    </div>

    <div v-if="isQuiz" class="modal is-active">
      <div class="modal-background"></div>
      <div class="modal-card">
        <header class="modal-card-head">
          <p class="modal-card-title">Quiz</p>
          <button
            class="delete"
            aria-label="close"
            @click="
              () => {
                isQuiz = false
                doFilter()
              }
            "
          ></button>
        </header>
        <section class="modal-card-body">
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
import { ref, watch, defineComponent, onBeforeMount, nextTick } from 'vue'

import { api } from '@/assets/api'

import Quiz from './Quiz.vue'
import LeechContent from './LeechContent.vue'

export default defineComponent({
  components: {
    Quiz,
    LeechContent
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
      next: ''
    })
    const isLeechOpen = ref(false)
    const isQuiz = ref(false)
    const sessionId = ref('')

    const doFilter = () => {
      api
        .get<{
          new: number
          due: number
          leech: number
          next: string
        }>('/api/quiz/stat', {
          params: {
            q: q.value,
            state: state.value.join(',')
          }
        })
        .then(({ data }) => {
          stat.value = data

          if (isLeechOpen.value) {
            isLeechOpen.value = false
            nextTick(() => {
              isLeechOpen.value = true
            })
          }
        })
    }

    const doQuiz = () => {
      api
        .post<{
          id: string
          count: number
        }>('/api/quiz/init', undefined, {
          params: {
            q: q.value,
            state: state.value.join(',')
          }
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
      isLeechOpen,
      isQuiz,
      sessionId
    }
  }
})
</script>

<style lang="scss" scoped>
#State {
  display: grid;
  grid-template-columns: minmax(80px, 1.5fr) 1fr 1fr;
  gap: 1em;
  max-width: 800px;
  margin: 0 auto;

  .field:not(:first-child) {
    justify-self: end;
    padding-right: 1em;
  }

  .check-list {
    display: grid;
    grid-auto-flow: column;
    gap: 1em;

    @media screen and (max-width: 630px) {
      grid-auto-flow: row;
    }
  }
}

#Stat {
  .card-content {
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
      margin-left: 0.5em;
    }
  }
}

#Leech {
  .clickable {
    cursor: pointer;
  }
}

.modal {
  .modal-card {
    width: 600px;
    max-width: 1200px;
    height: 100%;
  }

  .modal-card-body {
    height: 100%;
  }
}
</style>
