<template>
    <form v-if="service.type" @submit.prevent="saveService">
        <div class="card contain-card mb-4">
            <div class="card-header">{{ $t('service_info') }}</div>
            <div class="card-body">
            <div class="form-group row">
                <label class="col-sm-4 col-form-label">{{ $t('service_name') }}</label>
                <div class="col-sm-8">
                    <input v-model="service.name" @input="updatePermalink" id="name" type="text" name="name" class="form-control" placeholder="Server Name" required spellcheck="false" autocorrect="off">
                    <small class="form-text text-muted">Give your service a name you can recognize</small>
                </div>
            </div>
        <div class="form-group row">
            <label for="service_type" class="col-sm-4 col-form-label">{{ $t('service_type') }}</label>
            <div class="col-sm-8">
                <select v-model="service.type" @change="updateDefaultValues()" class="form-control" id="service_type">
                    <option value="http">HTTP {{ $t('service') }}</option>
                    <option value="tcp">TCP {{ $t('service') }}</option>
                    <option value="udp">UDP {{ $t('service') }}</option>
                    <option value="icmp">ICMP Ping</option>
                    <option value="grpc">gRPC {{ $t('service') }}</option>
                    <option value="ssh">SSH {{ $t('service') }}</option>
                    <option value="static">Static {{ $t('service') }}</option>
                </select>
                <small class="form-text text-muted">Use HTTP if you are checking a website or use TCP if you are checking a server</small>
            </div>
        </div>
        <div class="form-group row">
            <label for="service_type" class="col-sm-4 col-form-label">{{ $t('group') }}</label>
            <div class="col-sm-8">
                <select v-model.number="service.group_id" class="form-control">
                    <option value="0" >No Group</option>
                    <option v-for="(group, index) in $store.getters.cleanGroups()" :value="group.id">{{group.name}}</option>
                </select>
                <small class="form-text text-muted">Attach this service to a group</small>
            </div>
        </div>
            <div class="form-group row">
                <label class="col-sm-4 col-form-label">{{ $t('permalink') }}</label>
                <div class="col-sm-8">
                    <input v-model="service.permalink" type="text" name="permalink" class="form-control" id="permalink" autocapitalize="none" spellcheck="true" placeholder='awesome_service'>
                    <small class="form-text text-muted">Use text for the service URL rather than the service number.</small>
                </div>
            </div>

            <div class="form-group row">
                <label class="col-sm-4 col-form-label">{{ $t('service_public') }}</label>
                <div class="col-12 col-md-8 mt-1 mb-2">
                    <span @click="service.public = !!service.public" class="switch float-left">
                        <input v-model="service.public" type="checkbox" name="public-option" class="switch" id="switch-public" v-bind:checked="service.public">
                        <label v-if="service.public" for="switch-public">This service will be visible for everyone</label>
                        <label v-if="!service.public" for="switch-public">This service will only be visible for users and administrators.</label>
                    </span>
                </div>
            </div>

            <div v-if="service.type !== 'static'" class="form-group row">
                <label for="service_interval" class="col-sm-4 col-form-label">{{ $t('check_interval') }}</label>
                <div class="col-sm-6">
                    <span class="slider-info">{{secondsHumanize(service.check_interval)}}</span>
                    <input v-model.number="service.check_interval" type="range" class="slider" id="service_interval" min="1" max="1800" :step="1">
                    <small id="interval" class="form-text text-muted">Interval to check your service state</small>
                </div>
                <div class="col-sm-2">
                    <input v-model.number="service.check_interval" type="number" name="check_interval" class="form-control">
                </div>
            </div>

            </div>
        </div>

        <div v-if="service.type !== 'static'" class="card contain-card mb-4">
            <div class="card-header">Request Details</div>
            <div class="card-body">

            <div class="form-group row">
                <label for="service_url" class="col-sm-4 col-form-label">
                  {{ $t('service_endpoint') }} {{service.type === 'http' ? "(URL)" : "(Domain)"}}
                </label>
                <div class="col-sm-8">
                    <input v-model="service.domain" type="url" class="form-control" id="service_url" :placeholder="service.type === 'http' ? 'https://google.com' : '192.168.1.1'" required autocapitalize="none" spellcheck="false">
                    <small class="form-text text-muted">Statping will attempt to connect to this address</small>
                </div>
            </div>

            <div v-if="service.type.match(/^(tcp|udp|grpc|ssh)$/)" class="form-group row">
                <label class="col-sm-4 col-form-label">Port</label>
                <div class="col-sm-8">
                    <input v-model.number="service.port" type="number" name="port" class="form-control" id="service_port" placeholder="8080">
                </div>
            </div>

            <div v-if="service.type.match(/^(http)$/)" class="form-group row">
                <label class="col-sm-4 col-form-label">{{ $t('service_check') }}</label>
                <div class="col-sm-8">
                    <select v-model="service.method" name="method" class="form-control">
                        <option value="GET" >GET</option>
                        <option value="POST" >POST</option>
                        <option value="DELETE" >DELETE</option>
                        <option value="PATCH" >PATCH</option>
                        <option value="PUT" >PUT</option>
                    </select>
                    <small class="form-text text-muted">A GET request will simply request the endpoint, you can also send data with POST.</small>
                </div>
            </div>

        <div class="form-group row">
            <label class="col-sm-4 col-form-label">{{ $t('service_timeout') }}</label>
            <div class="col-sm-6">
                <span v-if="service.timeout >= 0" class="slider-info">{{secondsHumanize(service.timeout)}}</span>
                <input v-model.number="service.timeout" type="range" id="timeout" name="timeout" class="slider" min="1" max="180">
                <small class="form-text text-muted">If the endpoint does not respond within this time it will be considered to be offline</small>
            </div>

            <div class="col-sm-2">
                <input v-model.number="service.timeout" type="number" name="service_timeout" class="form-control">
            </div>

        </div>

        <div v-if="service.type.match(/^(http)$/) && service.method.match(/^(POST|PATCH|DELETE|PUT)$/)" class="form-group row">
            <label class="col-sm-4 col-form-label">Optional Post Data (JSON)</label>
            <div class="col-sm-8">
                <textarea v-model="service.post_data" class="form-control" rows="3" autocapitalize="none" spellcheck="false" placeholder='{"data": { "method": "success", "id": 148923 } }'></textarea>
                <small class="form-text text-muted">Insert a JSON string to send data to the endpoint.</small>
            </div>
        </div>
        <div v-if="service.type.match(/^(http)$/)" class="form-group row">
            <label class="col-sm-4 col-form-label">HTTP Headers</label>
            <div class="col-sm-8">
                <input v-model="service.headers" class="form-control" autocapitalize="none" spellcheck="false" placeholder='Authorization=1010101,Content-Type=application/json'>
                <small class="form-text text-muted">Comma delimited list of HTTP Headers (KEY=VALUE,KEY=VALUE)</small>
            </div>
        </div>
        <div v-if="service.type.match(/^(http)$/)" class="form-group row">
            <label class="col-sm-4 col-form-label">{{ $t('expected_resp') }} (Regex)</label>
            <div class="col-sm-8">
                <textarea v-model="service.expected" class="form-control" rows="3" autocapitalize="none" spellcheck="false" placeholder='(method)": "((\\"|[success])*)"'></textarea>
                <small class="form-text text-muted">You can use plain text or insert <a target="_blank" href="https://regex101.com/r/I5bbj9/1">Regex</a> to validate the response</small>
            </div>
        </div>
        <div v-if="service.type.match(/^(http)$/)" class="form-group row">
            <label for="service_response_code" class="col-sm-4 col-form-label">{{ $t('expected_code') }}</label>
            <div class="col-sm-8">
                <input v-model="service.expected_status" type="number" name="expected_status" class="form-control" placeholder="200" id="service_response_code">
                <small class="form-text text-muted">A status code of 200 is success, or view all the <a target="_blank" href="https://www.restapitutorial.com/httpstatuscodes.html">HTTP Status Codes</a></small>
            </div>
        </div>

        <div v-if="service.type.match(/^(http)$/)" class="form-group row">
            <label class="col-12 col-md-4 col-form-label">{{ $t('follow_redir') }}</label>
            <div class="col-12 col-md-8 mt-1 mb-2 mb-md-0">
                <span @click="service.redirect = !!service.redirect" class="switch float-left">
                    <input v-model="service.redirect" type="checkbox" name="redirect-option" class="switch" id="switch-redirect" v-bind:checked="service.redirect">
                    <label for="switch-redirect">Follow HTTP Redirects if server attempts</label>
                </span>
            </div>
        </div>
        <div v-if="service.type.match(/^(http|grpc)$/)" class="form-group row">
            <label class="col-12 col-md-4 col-form-label">{{ $t('verify_ssl') }}</label>
            <div class="col-12 col-md-8 mt-1 mb-2 mb-md-0">
                <span @click="service.verify_ssl = !!service.verify_ssl" class="switch float-left">
                    <input v-model="service.verify_ssl" type="checkbox" name="verify_ssl-option" class="switch" id="switch-verify-ssl" v-bind:checked="service.verify_ssl">
                    <label for="switch-verify-ssl" v-if="service.verify_ssl">Verify SSL Certificate for this service</label>
                    <label for="switch-verify-ssl" v-if="!service.verify_ssl">Skip SSL Certificate verification for this service</label>
                </span>
            </div>
        </div>

        <div v-if="service.type.match(/^(grpc)$/)" class="form-group row">
            <label class="col-12 col-md-4 col-form-label"><a href="https://github.com/grpc/grpc/blob/master/doc/health-checking.md#grpc-health-checking-protocol">GRPC Health Check</a></label>
            <div class="col-12 col-md-8 mt-1 mb-2 mb-md-0">
                <span @click="service.grpc_health_check = !!service.grpc_health_check" class="switch float-left">
                    <input v-model="service.grpc_health_check" type="checkbox" name="grpc_health_check-option" class="switch" id="switch-grpc-health-check" v-bind:checked="service.grpc_health_check">
                    <label for="switch-grpc-health-check" v-if="service.grpc_health_check">Check against GRPC health check endpoint.</label>
                    <label for="switch-grpc-health-check" v-if="!service.grpc_health_check">Only checks if GRPC connection can be established.</label>
                </span>
            </div>
        </div>

        <div v-if="service.grpc_health_check" class="form-group row">
            <label class="col-sm-4 col-form-label">Expected Response</label>
            <div class="col-sm-8">
                <textarea v-model="service.expected" class="form-control" rows="3" autocapitalize="none" spellcheck="false" placeholder='status:SERVING'></textarea>
                <small class="form-text text-muted">Check <a target="_blank" href="https://pkg.go.dev/google.golang.org/grpc/health/grpc_health_v1?tab=doc#pkg-variables">GPRC health check response codes</a> for more information.</small>
            </div>
        </div>

        <div v-if="service.grpc_health_check" class="form-group row">
            <label for="service_response_code" class="col-sm-4 col-form-label">Expected Status Code</label>
            <div class="col-sm-8">
                <input v-model="service.expected_status" type="number" name="expected_status" class="form-control" placeholder="1" id="service_response_code">
                <small class="form-text text-muted">A status code of 1 is success, or view all the <a target="_blank" href="https://pkg.go.dev/google.golang.org/grpc/health/grpc_health_v1?tab=doc#HealthCheckResponse_ServingStatus">GRPC Status Codes</a></small>
            </div>
        </div>

        <div v-if="service.type.match(/^(ssh)$/)" class="form-group row">
            <label for="service_username" class="col-sm-4 col-form-label">{{ $t('username') }}</label>
            <div class="col-sm-8">
                <input v-model="service.username" type="text" name="service_username" class="form-control" id="service_username">
            </div>
        </div>

        <div v-if="service.type.match(/^(ssh)$/)" class="form-group row">
            <label for="service_password" class="col-sm-4 col-form-label">{{ $t('password') }}</label>
            <div class="col-sm-8">
                <input v-model="service.password" type="text" name="service_password" class="form-control" id="service_password">
            </div>
        </div>

        <div v-if="service.type.match(/^(ssh)$/)" class="form-group row">
            <label class="col-12 col-md-4 col-form-label">SSH Health Check</label>
            <div class="col-12 col-md-8 mt-1 mb-2 mb-md-0">
                <span @click="service.ssh_health_check = !!service.ssh_health_check" class="switch float-left">
                    <input v-model="service.ssh_health_check" type="checkbox" name="ssh_health_check-option" class="switch" id="switch-ssh-health-check" v-bind:checked="service.ssh_health_check">
                    <label for="switch-ssh-health-check" v-if="service.ssh_health_check">Check against SSH health check command.</label>
                    <label for="switch-ssh-health-check" v-if="!service.ssh_health_check">Only checks if SSH connection can be established.</label>
                </span>
            </div>
        </div>

        <div v-if="service.ssh_health_check" class="form-group row">
            <label class="col-sm-4 col-form-label">Expected Response</label>
            <div class="col-sm-8">
                <textarea v-model="service.expected" class="form-control" rows="3" autocapitalize="none" spellcheck="false"></textarea>
                <small class="form-text text-muted">Only stdout will be parsed</small>
            </div>
        </div>

        <div v-if="service.ssh_health_check" class="form-group row">
            <label for="service_response_code" class="col-sm-4 col-form-label">Expected Exit Code</label>
            <div class="col-sm-8">
                <input v-model="service.expected_status" type="number" name="expected_status" class="form-control" placeholder="1" id="service_response_code">
                <small class="form-text text-muted">By convention a command exit code 0 indicates success</small>
            </div>
        </div>

        <div v-if="service.type.match(/^(ssh)$/) && service.ssh_health_check" class="form-group row">
            <label for="check_command" class="col-sm-4 col-form-label">Check Command</label>
            <div class="col-sm-8">
                <input v-model="service.check_command" type="text" name="check_command" class="form-control" id="check_command">
                <small class="form-text text-muted">Check command to be executed on remote server</small>
            </div>
        </div>

        <div v-if="service.type.match(/^(tcp|http)$/)" class="form-group row">
            <label class="col-12 col-md-4 col-form-label">{{ $t('tls_cert') }}</label>
            <div class="col-12 col-md-8 mt-1 mb-2 mb-md-0">
                <span @click="use_tls = !!use_tls" class="switch float-left">
                    <input v-model="use_tls" type="checkbox" name="verify_ssl-option" class="switch" id="switch-use-tls" v-bind:checked="use_tls">
                    <label for="switch-use-tls" v-if="use_tls">Custom TLS Certificates for mTLS services</label>
                    <label for="switch-use-tls" v-if="!use_tls">Ignore TLS Certificates</label>
                </span>
            </div>
        </div>

                <div v-if="use_tls" class="form-group row">
                    <label for="service_tls_cert" class="col-sm-4 col-form-label">TLS Client Certificate</label>
                    <div class="col-sm-8">
                        <textarea v-model="service.tls_cert" name="tls_cert" class="form-control" id="service_tls_cert"></textarea>
                        <small class="form-text text-muted">Absolute path to TLS Client Certificate file or in PEM format</small>
                    </div>
                </div>

                <div v-if="use_tls" class="form-group row">
                    <label for="service_tls_cert_key" class="col-sm-4 col-form-label">TLS Client Key</label>
                    <div class="col-sm-8">
                        <textarea v-model="service.tls_cert_key" name="tls_cert_key" class="form-control" id="service_tls_cert_key"></textarea>
                        <small class="form-text text-muted">Absolute path to TLS Client Key file or in PEM format</small>
                    </div>
                </div>

                <div v-if="use_tls" class="form-group row">
                    <label for="service_tls_cert_chain" class="col-sm-4 col-form-label">Root CA</label>
                    <div class="col-sm-8">
                        <textarea v-model="service.tls_cert_root" name="tls_cert_key" class="form-control" id="service_tls_cert_chain"></textarea>
                        <small class="form-text text-muted">Absolute path to Root CA file or in PEM format (optional)</small>
                    </div>
                </div>

            </div>
        </div>

        <div class="card contain-card mb-4">
            <div class="card-header">{{ $t('notification_opts') }}</div>
            <div class="card-body">

                <div class="form-group row">
                    <label class="col-sm-4 col-form-label">{{ $t('notifications_enable') }}</label>
                    <div class="col-12 col-md-8 mt-1 mb-2 mb-md-0">
                        <span @click="service.allow_notifications = !!service.allow_notifications" class="switch float-left">
                            <input v-model="service.allow_notifications" type="checkbox" name="allow_notifications-option" class="switch" id="switch-notifications" v-bind:checked="service.allow_notifications">
                            <label for="switch-notifications">Allow notifications to be sent for this service</label>
                        </span>
                    </div>
                </div>
                <div v-if="service.allow_notifications"  class="form-group row">
                    <label class="col-sm-4 col-form-label">{{ $t('notify_after') }}</label>
                    <div class="col-sm-8">
                        <span class="slider-info">{{service.notify_after === 0 ? "First Failure" : service.notify_after+' Failures'}}</span>
                        <input v-model="service.notify_after" type="range" name="notify_after" class="slider" id="notify_after" min="0" max="20">
                        <small class="form-text text-muted">Send Notification after {{service.notify_after === 0 ? 'the first Failure' : service.notify_after+' Failures'}} </small>
                    </div>
                </div>
                <div v-if="service.allow_notifications" class="form-group row">
                    <label class="col-sm-4 col-form-label">{{ $t('notify_all') }}</label>
                    <div class="col-12 col-md-8 mt-1">
                        <span @click="service.notify_all_changes = !!service.notify_all_changes" class="switch float-left">
                            <input v-model="service.notify_all_changes" type="checkbox" name="notify_all-option" class="switch" id="notify_all" v-bind:checked="service.notify_all_changes">
                            <label v-if="service.notify_all_changes" for="notify_all">Continuously send notifications when service is failing.</label>
                            <label v-if="!service.notify_all_changes" for="notify_all">Only notify one time when service hits an error</label>
                        </span>
                    </div>
                </div>

            </div>
        </div>

        <div class="form-group row">
            <div class="col-12">
                <button :disabled="loading" @click.prevent="saveService" type="submit" class="btn btn-success btn-block">
                    {{service.id ? $t('service_update') : $t('service_create')}}
                </button>
            </div>
        </div>

        <div class="alert alert-danger d-none" id="alerter" role="alert"></div>
    </form>
</template>

<script>
  import Api from "../API";

  export default {
      name: 'FormService',
      data () {
          return {
              loading: false,
              service: {
                  name: "",
                  type: "http",
                  domain: "",
                  group_id: 0,
                  method: "GET",
                  post_data: "",
                  headers: "",
                  expected: "",
                  expected_status: 200,
                  port: 80,
                  check_interval: 60,
                  timeout: 15,
                  permalink: "",
                  order: 1,
                  verify_ssl: true,
                  grpc_health_check: false,
                  ssh_health_check: false,
                  username: "",
                  password: "",
                  check_command: "",
                  redirect: true,
                  allow_notifications: true,
                  notify_all_changes: true,
                  notify_after: 2,
                  public: true,
                  tls_cert: "",
                  tls_cert_key: "",
                  tls_cert_root: "",
              },
              use_tls: false,
              groups: [],
          }
      },
      props: {
          in_service: {
              type: Object
          }
      },
      watch: {
          in_service(svr, old) {
            this.service = svr
            this.use_tls = svr.tls_cert
          }
      },
      async mounted () {
          if (!this.$store.getters.groups) {
            const groups = await Api.groups()
            this.$store.commit('setGroups', groups)
          }
        this.update()
      },
    created () {
        this.update()
    },
    methods: {
        update() {
          if (this.in_service) {
            this.service = this.in_service
          }
          this.use_tls = this.service.tls_cert !== ""
        },
        updateDefaultValues() {
            if (this.service.type === "grpc") {
                this.service.expected_status = 1
                this.service.expected = "status:SERVING"
                this.service.port = 50051
                this.service.verify_ssl = false
                this.service.method = ""
            } else if (this.service.type == "ssh") {
                this.service.port = 22;
                this.service.expected_status = 0;
            } else {
                this.service.expected_status = 200
                this.service.expected = ""
                this.service.port = 80
                this.service.verify_ssl = true
                this.service.method = "GET"
            }
        },
          updatePermalink() {
              const a = 'àáâäæãåāăąçćčđďèéêëēėęěğǵḧîïíīįìłḿñńǹňôöòóœøōõőṕŕřßśšşșťțûüùúūǘůűųẃẍÿýžźż·/_,:;'
              const b = 'aaaaaaaaaacccddeeeeeeeegghiiiiiilmnnnnoooooooooprrsssssttuuuuuuuuuwxyyzzz------'
              const p = new RegExp(a.split('').join('|'), 'g')

              this.service.permalink = this.service.name.toLowerCase()
                  .replace(/\s+/g, '-') // Replace spaces with -
                  .replace(p, c => b.charAt(a.indexOf(c))) // Replace special characters
                  .replace(/&/g, '-and-') // Replace & with 'and'
                  .replace(/[^\w\-]+/g, '') // Remove all non-word characters
                  .replace(/\-\-+/g, '-') // Replace multiple - with single -
                  .replace(/^-+/, '') // Trim - from start of text
                  .replace(/-+$/, '') // Trim - from end of text
          },
          stepVal(val) {
              if (val > 1800) {
                  return 300
              } else if (val > 300) {
                  return 60
              } else if (val > 120) {
                  return 10
              }
              return 1
          },
          async saveService () {
              let s = this.service
              this.loading = true
              delete s.failures
              delete s.created_at
              delete s.updated_at
              delete s.last_success
              delete s.latency
              delete s.online_24_hours
              s.check_interval = parseInt(s.check_interval)
              s.timeout = parseInt(s.timeout)
              s.port = parseInt(s.port)
              s.notify_after = parseInt(s.notify_after)
              s.expected_status = parseInt(s.expected_status)
              s.order = parseInt(s.order)

              if (s.id) {
                  await this.updateService(s)
              } else {
                  await this.createService(s)
              }
              const services = await Api.services()
              this.$store.commit('setServices', services)
              this.loading = false
              this.$router.push('/dashboard/services')
          },
          async createService (s) {
              await Api.service_create(s)
          },
          async updateService (s) {
              await Api.service_update(s)
          }
      }
  }
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
</style>
