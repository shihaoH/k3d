/*
Copyright © 2019 Thorsten Klein <iwilltry42@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cluster

import (
	"bytes"
	"io/ioutil"
	"os"

	"github.com/rancher/k3d/pkg/runtimes"
	k3d "github.com/rancher/k3d/pkg/types"
	log "github.com/sirupsen/logrus"
)

// GetKubeconfig grabs the kubeconfig file from /output from a master node container and puts it into a local directory
func GetKubeconfig(runtime runtimes.Runtime, cluster *k3d.Cluster) ([]byte, error) {
	masterNodes, err := runtime.GetNodesByLabel(map[string]string{"k3d.cluster": cluster.Name, "k3d.role": string(k3d.MasterRole)})
	if err != nil {
		log.Errorln("Failed to get masternodes")
		return nil, err
	}
	reader, err := runtime.GetKubeconfig(masterNodes[0])
	if err != nil {
		log.Errorf("Failed to get kubeconfig from node '%s'", masterNodes[0].Name)
		return nil, err
	}
	defer reader.Close()

	readBytes, err := ioutil.ReadAll(reader)
	if err != nil {
		log.Errorln("Couldn't read kubeconfig file")
		return nil, err
	}

	// write to file, skipping the first 512 bytes which contain file metadata
	// and trimming any NULL characters
	trimBytes := bytes.Trim(readBytes[512:], "\x00")

	return trimBytes, nil
}

// GetKubeconfigPath uses GetKubeConfig to grab the kubeconfig from the cluster master node, writes it to a file and outputs the path
func GetKubeconfigPath(runtime runtimes.Runtime, cluster *k3d.Cluster, path string) (string, error) {
	var output *os.File
	defer output.Close()
	var err error

	kubeconfigBytes, err := GetKubeconfig(runtime, cluster)
	if err != nil {
		log.Errorln("Failed to get kubeconfig")
		return "", err
	}

	if path == "-" {
		output = os.Stdout
	} else {
		if path == "" {
			path = "/tmp/test.yaml" // TODO: set proper default
		}
		output, err = os.Create(path)
		if err != nil {
			log.Errorf("Failed to create file '%s'", path)
			return "", err
		}
	}

	_, err = output.Write(kubeconfigBytes)
	if err != nil {
		log.Errorf("Failed to write to file '%s'", output.Name())
		return "", err
	}

	return output.Name(), nil

}