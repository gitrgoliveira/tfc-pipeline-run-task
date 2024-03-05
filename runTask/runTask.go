package runTask

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// downloadTarGz downloads a tar.gz file from the specified URL and extracts its contents.
//
// It takes a URL as a parameter and returns an error if any error occurs during the download or extraction process.
// The function creates a temporary file to store the downloaded tar.gz file, sends a GET request to download the file,
// copies the response body to the temporary file, opens the downloaded tar.gz file for reading, creates a reader for
// the gzip file, and extracts files from the tar archive. It returns nil if the process is successful.
func DownloadConfigVersion(url, token, folder string) error {
	// Create a new HTTP request with the provided URL
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	// Add the token to the request header
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/vnd.api+json")

	// Send the request
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check if the response status code is 200 OK
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed with status code: %d", resp.StatusCode)
	}

	// Create a temporary file to store the downloaded tar.gz file
	tarFilePath := filepath.Join(folder, folder+".tar.gz")
	tarFile, err := os.Create(tarFilePath)
	if err != nil {
		return err
	}
	defer os.Remove(tarFile.Name())

	// Copy the response body to the temporary file
	_, err = io.Copy(tarFile, resp.Body)
	if err != nil {
		return err
	}

	// Close the temporary file
	err = tarFile.Close()
	if err != nil {
		return err
	}
	log.Println("download OK")

	err = extractTarGz(tarFile.Name(), "./"+folder)
	if err != nil {
		log.Println(err.Error())
	} else {
		log.Println("File extracted OK")
	}

	return nil

}

func extractTarGz(tarGzFile, destination string) error {
	// Open the tar.gz file for reading
	file, err := os.Open(tarGzFile)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create a gzip reader for the file
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	// Create a tar reader for the gzip reader
	tarReader := tar.NewReader(gzipReader)

	// Extract files from the tar archive
	for {
		header, err := tarReader.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		// Extract the file to the specified destination
		target := filepath.Join(destination, header.Name)

		// Check the type of entry
		switch header.Typeflag {
		case tar.TypeDir:
			// Create directory if it doesn't exist
			err = os.MkdirAll(target, os.ModePerm)
			if err != nil {
				return err
			}

		case tar.TypeReg:
			// Create the file and copy contents
			err = os.MkdirAll(filepath.Dir(target), os.ModePerm)
			if err != nil {
				return err
			}

			file, err := os.Create(target)
			if err != nil {
				return err
			}
			defer file.Close()

			if _, err := io.Copy(file, tarReader); err != nil {
				return err
			}

		default:
			return fmt.Errorf("unsupported file type: %v in %s", header.Typeflag, header.Name)
		}
	}

	return nil
}

func DownloadPlan(url string, authToken string, filePath string) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+authToken)

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil || resp == nil {
		return err
	}

	defer resp.Body.Close()

	// Create the plan file
	planFile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer planFile.Close()

	// Copy the response body to the plan file
	_, err = io.Copy(planFile, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
