<template>
<div>
  <h3>Configuration</h3>
  For security reasons, all database credentials cannot be edited from this page.

  <codemirror v-show="loaded" v-model="configs" ref="configs" :options="cmOptions" class="mt-4 codemirrorInput"/>

  <button @click.prevent="save" class="btn col-12 btn-primary mt-3">Save</button>
</div>
</template>

<script>
import Api from "../../API";

import {codemirror} from 'vue-codemirror'
import('codemirror/lib/codemirror.css')
import('codemirror/mode/yaml/yaml.js')

export default {
name: "Configs",
  components: {
    codemirror
  },
  data() {
    return {
      loaded: false,
      configs: null,
      cmOptions: {
        height: 700,
        tabSize: 4,
        lineNumbers: true,
        matchBrackets: true,
        mode: "text/x-yaml",
        line: true
      }
    }
  },
  mounted() {
    this.loaded = false
    this.update()
    this.loaded = true
  },
  watch: {
    "configs" () {
      this.$refs.configs.codemirror.refresh()
    }
  },
  methods: {
    async update() {
      this.configs = await Api.configs()
      this.$refs.configs.codemirror.value = this.configs
      this.$refs.configs.codemirror.refresh()
    },
    async save() {
      try {
        await Api.configs_save(this.configs)
      } catch(e) {
        window.console.error(e)
      }
    }
  }
}
</script>

<style scoped>

</style>
