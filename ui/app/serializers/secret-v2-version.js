import { get } from '@ember/object';
import { assign } from '@ember/polyfills';
import ApplicationSerializer from './application';

export default ApplicationSerializer.extend({
  secretDataPath: 'data.data',
  normalizeItems(payload) {
    let path = this.secretDataPath;
    // move response that is the contents of the secret from the dataPath
    // to `secret_data` so it will be `secretData` in the model
    payload.secret_data = get(payload, path);
    payload = assign({}, payload, payload.data.metadata);
    delete payload.data;
    // payload.path = payload.id; // ARG TODO I'm not sure this does anything.
    // return the payload if it's expecting a single object or wrap
    // it as an array if not
    return payload;
  },
  serialize(snapshot) {
    let secret = snapshot.belongsTo('secret'); // ARG TODO this is the error, it returns undefined or null or blank but only when changing the maxVersions. Likely because ID is not being set on teh secret-v2-version
    if (!secret) {
      console.log('!!!!! Failure !!!!!', snapshot); // THIS IS WHERE THE ERROR IS.
      return;
    }
    // if both models failed to read from the server, we need to write without CAS
    if (secret.record.failedServerRead && snapshot.record.failedServerRead) {
      return {
        data: snapshot.attr('secretData'),
      };
    }
    let version = secret.record.failedServerRead ? snapshot.attr('version') : secret.attr('currentVersion');
    version = version || 0;
    // ARG TODO This return is strange to me.
    return {
      data: snapshot.attr('secretData'),
      options: {
        cas: version,
      },
    };
  },
});
