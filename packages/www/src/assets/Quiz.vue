<template>
  <div id="Quiz">
    <iframe
      :src="`/card.html?side=${side}&id=${card.id}&secret=${secret}`"
      sandbox="allow-scripts allow-same-origin allow-forms"
    ></iframe>
    <footer>
      <button
        :disabled="!(index > 0)"
        :style="{ visibility: index > 0 ? 'visible' : 'hidden' }"
        class="button"
        type="button"
        @click="index--"
      >
        Previous
      </button>

      <div style="flex-grow: 1"></div>

      <button
        v-if="side !== 'back'"
        class="button is-warning"
        type="button"
        @click="side = 'back'"
      >
        Show answer
      </button>
      <button
        v-else
        class="button is-warning"
        type="button"
        @click="side = 'front'"
      >
        Hide answer
      </button>

      <button
        v-if="side !== 'front'"
        class="button is-primary"
        @click="dSrsLevel(1)"
      >
        Right
      </button>
      <button
        v-if="side !== 'front'"
        class="button is-danger"
        @click="dSrsLevel(-1)"
      >
        Wrong
      </button>
      <button
        v-if="side !== 'front'"
        class="button is-info"
        @click="dSrsLevel(0)"
      >
        Repeat
      </button>

      <button
        v-if="side !== 'front'"
        :class="`button ` + (card.isMarked ? 'is-warning' : 'is-success')"
        @click="toggleMark()"
      >
        {{ card.isMarked ? 'Unmark' : 'Mark' }}
      </button>

      <button
        v-if="side === 'back'"
        class="button has-background-grey-lighter"
        type="button"
        @click="side = 'mnemonic'"
      >
        Mnemonic
      </button>

      <div style="flex-grow: 1"></div>

      <button
        :disabled="!(index < cards.length - 2)"
        :style="{ visibility: index < cards.length - 2 ? 'visible' : 'hidden' }"
        class="button has-background-grey-lighter"
        type="button"
        @click="index++"
      >
        Next
      </button>

      <button
        v-if="index >= cards.length - 2"
        class="button is-success"
        type="button"
        @click="() => $emit('end') && endQuiz()"
      >
        End Quiz
      </button>
    </footer>
  </div>
</template>

<script lang="ts">
import { defineComponent, onMounted, ref } from 'vue'
import { api } from './api'

export default defineComponent({
  props: ['session'],
  emits: ['end'],
  setup() {
    const side = ref('front')
    const index = ref(0)
    const cards = ref(
      [] as {
        id: string
        dSrsLevel?: number
        isMarked?: boolean
      }[]
    )

    const dSrsLevel = (df: number) => {
      const c = cards.value[index.value]
      c.dSrsLevel = df
      cards.value = [
        ...cards.value.slice(0, index.value),
        c,
        ...cards.value.slice(index.value + 1),
      ]
    }

    const toggleMark = () => {
      const c = cards.value[index.value]
      c.isMarked = !c.isMarked
      cards.value = [
        ...cards.value.slice(0, index.value),
        c,
        ...cards.value.slice(index.value + 1),
      ]
    }

    const endQuiz = () => {
      console.log('Ending quiz')
    }

    onMounted(() => {
      api
        .get<{
          result: string[]
        }>('/api/quiz/leech', {
          params: {
            page: 1,
            limit: 10,
            q: '',
          },
        })
        .then(({ data }) => {
          cards.value = data.result.map((id) => ({ id }))
          index.value = 0
        })
    })

    return {
      cards,
      index,
      side,
      secret: new URL(location.href).searchParams.get('secret'),
      endQuiz,
      dSrsLevel,
      toggleMark,
    }
  },
  computed: {
    card(): {
      id: string
      dSrsLevel?: number
      isMarked?: boolean
    } {
      return this.cards[this.index] || {}
    },
  },
})
</script>

<style lang="scss">
#Quiz {
  display: grid;
  grid-template-rows: 1fr auto;
  height: 100%;
  width: 100%;

  iframe {
    height: 100%;
    width: 100%;
    border: none;
  }

  footer {
    display: flex;
    flex-direction: row;
    margin-top: 1em;
  }

  button + button {
    margin-left: 1em;
  }
}
</style>
