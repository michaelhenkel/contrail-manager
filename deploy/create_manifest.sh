cat role.yaml > 1-create-operator.yaml
echo "---" >> 1-create-operator.yaml
cat cluster_role.yaml >> 1-create-operator.yaml
echo "---" >> 1-create-operator.yaml
cat service_account.yaml >> 1-create-operator.yaml
echo "---" >> 1-create-operator.yaml
cat role_binding.yaml >> 1-create-operator.yaml
echo "---" >> 1-create-operator.yaml
cat cluster_role_binding.yaml >> 1-create-operator.yaml
echo "---" >> 1-create-operator.yaml
cat crds/contrail_v1alpha1_manager_crd.yaml >> 1-create-operator.yaml
echo "---" >> 1-create-operator.yaml
cat operator.yaml >> 1-create-operator.yaml

echo "---" > 2-start-operator-1node.yaml
cat crds/contrail_v1alpha1_manager_cr.yaml >> 2-start-operator-1node.yaml
echo "---" > 2-start-operator-3node.yaml
sed 's/size: "1"/size: "3"/g' 2-start-operator-1node.yaml > 2-start-operator-3node.yaml
