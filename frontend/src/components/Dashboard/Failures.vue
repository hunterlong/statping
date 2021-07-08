<template>
    <div v-if="service" class="col-12">
        <h3>{{service.name}} Failures
            <button v-if="failures.length>0" @click="deleteFailures" class="btn btn-danger float-right">Delete All</button>
        </h3>

        <div class="card mt-4 mb-4">
            <div class="card-header">
                Search and Filter
                <span class="float-right">
                    <font-awesome-icon v-if="loading" icon="circle-notch" spin/>
                </span>
            </div>
            <div class="card-body">
                <form>
                    <div class="form-row">
                        <div class="col">
                            <label for="fromdate">From Date</label>
                            <flatPickr id="fromdate" :disabled="loading" @on-change="load" v-model="start_time" :config="{ wrap: true, allowInput: true, enableTime: true, dateFormat: 'Z', altInput: true, altFormat: 'Y-m-d h:i K', maxDate: new Date() }" type="text" class="form-control text-left d-block" required />
                        </div>
                        <div class="col">
                            <label for="todate">To Date</label>
                            <flatPickr id="todate" :disabled="loading" @on-change="load" v-model="end_time" :config="{ wrap: true, allowInput: true, enableTime: true, dateFormat: 'Z', altInput: true, altFormat: 'Y-m-d h:i K', maxDate: new Date() }" type="text" class="form-control text-left d-block" required />
                        </div>
                        <div class="col">
                            <label for="search">Search Terms</label>
                            <input id="search" type="text" v-model="search" class="form-control">
                        </div>
                    </div>
                    <div class="form-row mt-3">
                        <div class="col">
                            <span @click="show_checkins = !!show_checkins" class="switch float-left">
                                <input v-model="show_checkins" type="checkbox" class="switch" id="showcheckins" v-bind:checked="show_checkins">
                                 <label v-if="show_checkins" for="showcheckins">Showing Checkin Failures</label>
                                 <label v-else for="showcheckins">View Checkin Failures</label>
                            </span>
                        </div>
                    </div>
                </form>
            </div>
        </div>

        <div v-if="failures.length === 0" class="alert alert-info">
            <span v-if="search">
                Could not find any failures with issue: "{{search}}"
            </span>
            <span v-else>
                You don't have any failures for {{service.name}}. Way to go!
            </span>
        </div>

        <table v-else class="table">
            <thead>
            <tr>
                <th scope="col">#</th>
                <th scope="col">Issue</th>
                <th scope="col">Status Code</th>
                <th scope="col">Ping</th>
                <th scope="col">Created (Day)</th>
                <th scope="col">Created (Hour)</th>
            </tr>
            </thead>
            <tbody>
            <tr v-for="(failure, index) in failures" :key="index">
                <th class="font-1" scope="row">{{failure.id}}</th>
                <td class="font-1">{{failure.issue}}</td>
                <td class="font-1">{{failure.error_code}}</td>
                <td class="font-1">{{humanTime(failure.ping)}}</td>
                <td class="font-1">{{formatDate(failure.created_at, 'yyyy-MM-dd')}}</td>
                <td class="font-1">{{formatDate(failure.created_at, 'hh:mm')}}</td>
            </tr>

            </tbody>
        </table>

        <nav v-if="total > 4 && failures.length !== 0" class="mt-3">
            <ul class="pagination justify-content-center">
                <li class="page-item" :class="{'disabled': page===1}">
                    <a @click.prevent="gotoPage(page-1)" :disabled="page===1" class="page-link" href="#" aria-label="Previous">
                        <span aria-hidden="true">&laquo;</span>
                        <span class="sr-only">Previous</span>
                    </a>
                </li>
                <li v-for="n in maxPages" class="page-item" :class="{'active': page === n}">
                    <a @click.prevent="gotoPage(n)" class="page-link" href="#">{{n}}</a>
                </li>
                <li class="page-item" :class="{'disabled': page===Math.floor(total / limit)}">
                    <a @click.prevent="gotoPage(page+1)" :disabled="page===Math.floor(total / limit)" class="page-link" href="#" aria-label="Next">
                        <span aria-hidden="true">&raquo;</span>
                        <span class="sr-only">Next</span>
                    </a>
                </li>
            </ul>
            <div class="text-center">
                <span>{{total}} Failures</span>
            </div>
        </nav>

    </div>
</template>

<script>
import Api from "../../API";
import flatPickr from 'vue-flatpickr-component';
import 'flatpickr/dist/flatpickr.css';

export default {
    name: 'Failures',
      components: {
        flatPickr
      },
    data() {
        return {
          loading: true,
          search: "",
          show_checkins: false,
          service: null,
          fails: [],
          limit: 64,
          offset: 0,
          total: 0,
          page: 1,
          start_time: this.nowSubtract(216000).toISOString(),
          end_time: this.nowSubtract(0).toISOString(),
        }
    },
      watch: {
        '$route': 'reloadTimes',
      },
      computed: {
        failures() {
          let sorted = this.fails
          if (this.show_checkins) {
            sorted = sorted.filter(f => f.method === "checkin");
          } else {
            sorted = sorted.filter(f => f.method !== "checkin");
          }
          if (this.search !== "") {
            sorted = sorted.filter(f => f.issue.toLowerCase().includes(this.search));
          }
          return sorted
        },
        pages() {
          return Math.floor(this.total / this.limit)
        },
        maxPages() {
          return Math.floor(this.total / this.limit)
        }
      },
      async created() {
        this.service = await Api.service(this.$route.params.id)
        this.total = this.service.stats.failures
        await this.gotoPage(1)
      },
    methods: {
      async delete() {
        await Api.service_failures_delete(this.service)
        this.service = await Api.service(this.service.id)
        this.total = 0
        await this.load()
      },
      async deleteFailures() {
        const modal = {
          visible: true,
          title: "Delete All Failures",
          body: `Are you sure you want to delete all Failures for service ${this.service.title}?`,
          btnColor: "btn-danger",
          btnText: "Delete Failures",
          func: () => this.delete(),
        }
        this.$store.commit("setModal", modal)
      },
      async gotoPage(page) {
        this.page = page;
        this.offset = (page-1) * this.limit;
        await this.load()
      },
      async load() {
        this.loading = true
        this.fails = await Api.service_failures(this.service.id, this.toUnix(this.parseISO(this.start_time)), this.toUnix(this.parseISO(this.end_time)), this.limit, this.offset)
        this.loading = false
      }
    }
}
</script>
