<template>
  <div id="machines-form">
    <h4>Create Machines</h4>
    <b-form class="mb-3" inline @submit="onSubmit">
      <b-form-select
        id="inline-form-custom-select-provider"
        class="mb-2 mr-sm-2 mb-sm-0"
        v-model="form.provider"
        :value="null"
        :options="providers"
        required
      >
      </b-form-select>

      <b-form-select
        id="inline-form-custom-select-regions"
        class="mb-2 mr-sm-2 mb-sm-0"
        v-model="form.region"
        :value="null"
        :options="regions"
        @change="onRegionChange"
      >
        <option slot="first" :value="null">Regions</option>
      </b-form-select>

      <b-form-select
        id="inline-form-custom-select-sizes"
        class="mb-2 mr-sm-2 mb-sm-0"
        v-model="form.size"
        :value="null"
        :options="sizes"
      >
        <option slot="first" :value="null">Size</option>
      </b-form-select>

      <!-- <b-form-select
        id="inline-form-custom-select-count"
        class="mb-2 mr-sm-2 mb-sm-0"
        v-model="form.count"
        :value="null"
        :options="count"
      >
        <option slot="first" :value="null">Count</option>
      </b-form-select> -->

      <b-form-input
        id="inline-form-custom-select-count"
        class="mb-2 mr-sm-2 mb-sm-0"
        v-model="form.count"
        type="number"
        :number="true"
        ></b-form-input>

      <b-button type="submit" variant="primary">Apply</b-button>
      <b-button class="delete-all" variant="danger" @click="deleteAllMachines">Delete All</b-button>
    </b-form>

    <b-alert
      :show="dismissCountDown"
      dismissible
      fade
      variant="success"
      @dismissed="dismissCountDown=0"
      @dismiss-count-down="countDownChanged"
    >
      Created Machine!
    </b-alert>
    <hr />
  </div>
</template>

<script>
export default {
  data() {
    return {
      form: {
        provider: 'gke',
        region: null,
        size: null,
        count: 0,
      },
      providers: { gke: 'media-processing' },
      regionsData: [],
      sizesData: [],
      sizes: [],
      count: Object.assign({}, [...Array(10).keys()]), // 0..10.
      dismissSecs: 5,
      dismissCountDown: 0,
      showDismissibleAlert: false,
    };
  },

  computed: {
    regions() {
      // return this.regionsData.map(o => o.name).sort();
      return this.regionsData.map((o) => {
        const obj = {
          text: `${o.name}`,
          value: o.slug,
        };
        return obj;
      }).sort();
    },
  },

  mounted() {
    this.getRegions();
    this.getSizes();
  },

  methods: {
    countDownChanged(dismissCountDown) {
      this.dismissCountDown = dismissCountDown;
    },

    getRegions() {
      const url = '/api/machines/regions';

      fetch(url)
        .then(response => (
          response.json()
        ))
        .then((json) => {
          this.regionsData = json && json.regions;
        });
    },

    getSizes() {
      const url = '/api/machines/sizes';

      fetch(url)
        .then(response => (
          response.json()
        ))
        .then((json) => {
          this.sizesData = json && json.sizes;
        });
    },

    createMachine(data) {
      const url = '/api/machines';

      fetch(url, {
        method: 'POST',
        body: JSON.stringify(data),
        headers: {
          'Content-Type': 'application/json',
        },
      }).then(response => (
        response.json()
      )).then((json) => {
        console.log('Created machine: ', json);
        this.dismissCountDown = this.dismissSecs;
      });
    },

    deleteAllMachines() {
      const url = '/api/machines/';

      fetch(url, {
        method: 'DELETE',
        headers: {
          'Content-Type': 'application/json',
        },
      }).then(response => (
        response.json()
      )).then((json) => {
        console.log('Deleting all machine: ', json);
      });
    },

    onRegionChange() {
      const { region } = this.form;

      // Get pricing data with regions available.
      const sizesAvailable = this.regionsData.find(o => o.slug === region).sizes;

      // Filter sizes data and sort by price to array.
      const filtered = this.sizesData
        .filter(o => sizesAvailable.includes(o.slug) && o.available)
        .sort((a, b) => a.price_monthly > b.price_monthly)
        .map((o) => {
          const obj = {
            text: `${o.slug} -- $${o.price_monthly} /mo`,
            value: o.slug,
          };
          return obj;
        });

      this.sizes = filtered;
    },

    onSubmit(evt) {
      evt.preventDefault();
      console.log(JSON.stringify(this.form));
      this.createMachine(this.form);
    },
  },
};
</script>

<style scoped>
#inline-form-custom-select-count {
  width: 80px;
}

button.delete-all {
  margin-left: auto;
}
</style>
