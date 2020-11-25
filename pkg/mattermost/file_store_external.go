package mattermost

//type ExternalFileStore struct {}
//
//func NewExternalFileStore(mattermost mattermostv1beta1.Mattermost, secret corev1.Secret) (*ExternalFileStore, error) {
//	if mattermost.Spec.Filestore.External == nil {
//		return nil, fmt.Errorf("external file store configuration not provided")
//	}
//
//	bucket := mattermost.Spec.Filestore.External.Bucket
//	if bucket == "" {
//		return nil, fmt.Errorf("external file store bucket is empty")
//	}
//
//	url := mattermost.Spec.Filestore.External.URL
//	if url == "" {
//		return nil, fmt.Errorf("external file store URL is empty")
//	}
//
//	if _, ok := secret.Data["accesskey"]; !ok {
//		return nil, fmt.Errorf("external filestore Secret %s does not have an 'accesskey' value", secret.Name)
//	}
//	if _, ok := secret.Data["secretkey"]; !ok {
//		return nil, fmt.Errorf("external filestore Secret %s does not have an 'secretkey' value", secret.Name)
//	}
//
//	return &ExternalFileStore{
//		secretName: secret.Name,
//		bucketName: bucket,
//		url: url,
//	}, nil
//}
//
//func (e *ExternalFileStore) Secret(mattermost *mattermostv1beta1.Mattermost) string {
//	return e.secretName
//}
//
//func (e *ExternalFileStore) Bucket(mattermost *mattermostv1beta1.Mattermost) string {
//	return e.bucketName
//}
//
//func (e *ExternalFileStore) URL() string {
//	return e.url
//}

//type ExternalFileStore struct {}
//
//func (e *ExternalFileStore) InitContainers(mattermost *mattermostv1beta1.Mattermost) []corev1.Container {
//	return []corev1.Container{}
//}

