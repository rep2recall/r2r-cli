<template>
  <div id="Quiz" style="display: grid; grid-template-rows: 1fr auto">
    <div
      v-if="side === 'mnemonic'"
      id="Mnemonic"
      style="display: flex; flex-direction: column"
    >
      <div class="quill"></div>
    </div>
    <iframe
      v-else-if="card.id"
      :src="`/card.html?side=${side}&id=${card.id}&secret=${secret}`"
      style="border-bottom: 1px solid rgba(128, 128, 128, 0.7)"
    ></iframe>
    <div v-else>
      <p>No quiz pending.</p>
    </div>

    <footer
      style="
        display: grid;
        grid-template-columns: 100px 1fr 100px;
        align-items: center;
        overflow: auto;
        min-height: 120px;
        max-height: 30vh;
      "
    >
      <div>
        <button
          :disabled="!(index > 0)"
          :style="{ visibility: index > 0 ? 'visible' : 'hidden' }"
          class="button"
          type="button"
          @click="index--"
        >
          Previous
        </button>
      </div>

      <div
        :style="{ visibility: card.id ? 'visible' : 'hidden' }"
        class="buttons"
      >
        <button
          v-if="side === 'front'"
          class="button is-warning"
          type="button"
          @click="side = 'back'"
        >
          Show answer
        </button>

        <div class="buttons">
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
        </div>

        <div class="buttons">
          <button
            v-if="side === 'back'"
            class="button is-warning"
            type="button"
            @click="side = 'front'"
          >
            Hide answer
          </button>
          <button
            v-if="side === 'mnemonic'"
            class="button is-warning"
            type="button"
            @click="side = 'back'"
          >
            Show answer
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
        </div>
      </div>

      <div style="margin-left: auto">
        <button
          v-if="index < cards.length - 2"
          class="button has-background-grey-lighter"
          type="button"
          @click="index++"
        >
          Next
        </button>

        <button
          v-else-if="side != 'front' && autoclose"
          class="button is-success"
          type="button"
          @click="() => $emit('end') && endQuiz()"
        >
          End Quiz
        </button>
      </div>
    </footer>
  </div>
</template>

<script lang="ts">
import { defineComponent, nextTick, onMounted, ref, watch } from 'vue'
import Quill from 'quill'
import { api } from './api'

import 'quill/dist/quill.snow.css'

let quill: Quill

export default defineComponent({
  props: ['session', 'standalone'],
  emits: ['end'],
  setup(props) {
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
      const i = index.value
      const c = cards.value[i]
      c.dSrsLevel = df

      api
        .patch('/api/card/dSrsLevel', undefined, {
          params: {
            id: c.id,
            dSrsLevel: c.dSrsLevel,
            session: props.session,
          },
        })
        .then(() => {
          cards.value = [
            ...cards.value.slice(0, i),
            c,
            ...cards.value.slice(i + 1),
          ]
          side.value = 'front'
          index.value = i + 1
        })
    }

    const toggleMark = () => {
      const i = index.value
      const c = cards.value[i]
      if (!c) {
        return
      }

      c.isMarked = !c.isMarked

      api
        .patch<{
          isMarked: boolean
        }>('/api/card/toggleMarked', undefined, {
          params: {
            id: c.id,
          },
        })
        .then(({ data }) => {
          c.isMarked = data.isMarked

          cards.value = [
            ...cards.value.slice(0, i),
            c,
            ...cards.value.slice(i + 1),
          ]
        })
    }

    const endQuiz = () => {
      // if (props.autoclose) {
      //   window.close()
      // }
    }

    watch(side, (side) => {
      const i = index.value
      const c = cards.value[i]

      if (side === 'mnemonic' && c) {
        nextTick(() => {
          if (quill) {
            quill.disable()
          }

          api
            .get('/api/card/mnemonic', {
              params: {
                id: c.id,
              },
            })
            .then((r) => {
              quill = new Quill('#Mnemonic .quill', {
                placeholder: 'Compose a memorable mnemonic...',
                theme: 'snow',
              })

              quill.setContents(r.data)
              quill.on('text-change', () => {
                console.log(quill.getContents())
                api.put('/api/card/mnemonic', quill.getContents(), {
                  params: {
                    id: c.id,
                  },
                })
              })

              r.data
            })
        })
      }
    })

    onMounted(() => {
      api
        .get<{
          result: {
            id: string
            isMarked: boolean
          }[]
        }>('/api/quiz/session', {
          params: {
            session: props.session,
          },
        })
        .then(({ data }) => {
          cards.value = data.result
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
      autoclose: !props.standalone,
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
  height: 100%;
  width: 100%;

  iframe {
    height: 100%;
    width: 100%;
    border: none;
  }

  button {
    margin: 0.5em;
  }

  .buttons {
    margin: 0 auto;
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
  }
}
</style>
