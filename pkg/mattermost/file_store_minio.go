package mattermost

//
//type OperatorManagedMinio struct {
//	secretName string
//	bucketName string
//	url string
//}
//
//func NewOperatorManagedMinio(mattermost mattermostv1beta1.Mattermost, secret corev1.Secret, minioURL string) (*OperatorManagedMinio, error) {
//
//	// TODO: do magic here
//
//	return &OperatorManagedMinio{
//		secretName: secret.Name,
//		bucketName: mattermost.Name,
//		url: minioURL,
//	}, nil
//}
//
//// TODO: do not pass Mattermost but save it before hand?
//
//func (e *OperatorManagedMinio) Secret(mattermost *mattermostv1beta1.Mattermost) string {
//	return fmt.Sprintf("%s-minio", mattermost.Name)
//}
//
//func (e *OperatorManagedMinio) Bucket(mattermost *mattermostv1beta1.Mattermost) string {
//	return mattermost.Name
//}
//
//func (e *OperatorManagedMinio) URL() string {
//	return e.url
//}

//type OperatorManagedMinio struct {
//	secretName string
//	bucketName string
//	url string
//}
//
//
//func (e *OperatorManagedMinio) InitContainers(mattermost *mattermostv1beta1.Mattermost) []corev1.Container {
//	return []corev1.Container{}
//}

