package readDir

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/lukemakhanu/magic_carpet/internal/domains/lsFiles"
	"github.com/lukemakhanu/magic_carpet/internal/domains/readFiles"
)

var _ readFiles.DirectoryReader = (*ReadDirectoryConfigs)(nil)

type ReadDirectoryConfigs struct {
	directory    string
	combinations string
}

// New initializes a new instance of Match Client.
func New(directory string, combinations string) (*ReadDirectoryConfigs, error) {

	if directory == "" {
		return nil, fmt.Errorf("directory not set")
	}

	if combinations == "" {
		return nil, fmt.Errorf("combinations not set")
	}

	c := &ReadDirectoryConfigs{
		directory:    directory,
		combinations: combinations,
	}

	return c, nil
}

// GetMatch : returns match and odds from third party provider
func (s *ReadDirectoryConfigs) ReadDirectory(ctx context.Context) ([]string, error) {
	fileList := []string{}
	f, err := os.Open(s.directory)
	if err != nil {
		return fileList, fmt.Errorf("Err : %v failed to read directory %s", err, s.directory)
	}

	files, err := f.Readdir(0)
	if err != nil {
		return fileList, fmt.Errorf("Err : %v failed to read files %s", err, s.directory)
	}

	allowed := strings.Split(s.combinations, ",")
	log.Println("Allowed combination ", allowed)

	for _, v := range files {

		log.Println(v.Name(), v.IsDir()) // 47461300_4_24.txt

		data := strings.Split(v.Name(), "_")
		if len(data) == 3 {

			chr := data[0]
			lastCharacter := chr[len(chr)-1:]

			log.Println("lastCharacter ", lastCharacter, "allowed", allowed)

			if contains(allowed, lastCharacter) == true {
				fileList = append(fileList, v.Name())
			} else {
				log.Println("Skip ", v.Name(), "allowed", allowed)
			}

		} else {
			log.Printf("Wrong data format : %s", v.Name())
		}

	}

	return fileList, nil
}

// AllFiles : returns all files in a directory
func (s *ReadDirectoryConfigs) AllFiles(ctx context.Context) ([]lsFiles.FileInfo, error) {
	fileList := []lsFiles.FileInfo{}
	f, err := os.Open(s.directory)
	if err != nil {
		return fileList, fmt.Errorf("Err : %v failed to read directory %s", err, s.directory)
	}

	files, err := f.Readdir(0)
	if err != nil {
		return fileList, fmt.Errorf("Err : %v failed to read files %s", err, s.directory)
	}

	allowed := strings.Split(s.combinations, ",")
	log.Println("Allowed combination ", allowed)

	for _, v := range files {

		log.Println(v.Name(), v.IsDir()) // 47461300_4_24.txt

		data := strings.Split(v.Name(), "_")
		if len(data) == 3 {

			extID := data[0]
			projectID := data[1]
			//competitionID := data[2]

			dat := strings.Split(data[2], ".")
			if len(dat) == 2 {

				rr := lsFiles.FileInfo{
					ExtID:         extID,
					ProjectID:     projectID,
					CompetitionID: dat[0],
					LsFileName:    v.Name(),
				}
				fileList = append(fileList, rr)

			}

		} else {
			log.Printf("Wrong data format : %s", v.Name())
		}

		//fileList = append(fileList, v.Name())

	}

	return fileList, nil
}

// SameRoundLiveScores : returns live scores for the round provided
func (s *ReadDirectoryConfigs) SameRoundLiveScores(ctx context.Context, fileNames []string) ([]string, error) {
	fileList := []string{}

	f, err := os.Open(s.directory)
	if err != nil {
		return fileList, fmt.Errorf("Err : %v failed to read directory %s", err, s.directory)
	}

	files, err := f.Readdir(0)
	if err != nil {
		return fileList, fmt.Errorf("Err : %v failed to read files %s", err, s.directory)
	}

	for _, v := range files {

		log.Println(v.Name(), v.IsDir())

		if contains(fileNames, v.Name()) == true {
			fileList = append(fileList, v.Name())
		} else {
			log.Println("Skip ", v.Name())
		}

	}

	return fileList, nil
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func (s *ReadDirectoryConfigs) ReadLiveScoreDirectory(ctx context.Context) ([]lsFiles.FileInfo, error) {
	fileList := []lsFiles.FileInfo{}
	f, err := os.Open(s.directory)
	if err != nil {
		return fileList, fmt.Errorf("Err : %v failed to read directory %s", err, s.directory)
	}

	files, err := f.Readdir(0)
	if err != nil {
		return fileList, fmt.Errorf("Err : %v failed to read files %s", err, s.directory)
	}

	allowed := strings.Split(s.combinations, ",")
	log.Println("Allowed combination ", allowed)

	for _, v := range files {

		log.Println(v.Name(), v.IsDir()) // 31114422_3_20.txt

		data := strings.Split(v.Name(), "_")
		if len(data) == 3 {

			extID := data[0]
			projectID := data[1]
			//competitionID := data[2]

			dat := strings.Split(data[2], ".")
			if len(dat) == 2 {

				rr := lsFiles.FileInfo{
					ExtID:         extID,
					ProjectID:     projectID,
					CompetitionID: dat[0],
					LsFileName:    v.Name(),
				}
				fileList = append(fileList, rr)

			}

		} else {
			log.Printf("Wrong data format : %s", v.Name())
		}

	}

	return fileList, nil
}

func (s *ReadDirectoryConfigs) ReadWinningOutcomeDirectory(ctx context.Context) ([]lsFiles.FileInfo, error) {
	fileList := []lsFiles.FileInfo{}
	f, err := os.Open(s.directory)
	if err != nil {
		return fileList, fmt.Errorf("Err : %v failed to read directory %s", err, s.directory)
	}

	files, err := f.Readdir(0)
	if err != nil {
		return fileList, fmt.Errorf("Err : %v failed to read files %s", err, s.directory)
	}

	allowed := strings.Split(s.combinations, ",")
	log.Println("Allowed combination ", allowed)

	for _, v := range files {

		log.Println(v.Name(), v.IsDir()) // wo_31114422_3_20.txt

		data := strings.Split(v.Name(), "_")
		if len(data) == 4 {

			extID := data[1]
			projectID := data[2]
			//competitionID := data[3]

			dat := strings.Split(data[3], ".")
			if len(dat) == 2 {

				rr := lsFiles.FileInfo{
					ExtID:         extID,
					ProjectID:     projectID,
					CompetitionID: dat[0],
					LsFileName:    v.Name(),
				}

				fileList = append(fileList, rr)
			}

		} else {
			log.Printf("Wrong data format : %s", v.Name())
		}

	}

	return fileList, nil
}

// ReadTeamDirectory : used to read team directory
func (s *ReadDirectoryConfigs) ReadTeamDirectory(ctx context.Context) ([]lsFiles.FileInfoTeams, error) {
	fileList := []lsFiles.FileInfoTeams{}
	f, err := os.Open(s.directory)
	if err != nil {
		return fileList, fmt.Errorf("Err : %v failed to read directory %s", err, s.directory)
	}

	files, err := f.Readdir(0)
	if err != nil {
		return fileList, fmt.Errorf("Err : %v failed to read files %s", err, s.directory)
	}

	allowed := strings.Split(s.combinations, ",")
	log.Println("Allowed combination ", allowed)

	for _, v := range files {

		log.Println(v.Name(), v.IsDir()) // 27196_2_380.txt

		data := strings.Split(v.Name(), "_")
		if len(data) == 3 {

			seasonID := data[0]
			competitionID := data[1]

			dat := strings.Split(data[2], ".")
			if len(dat) == 2 {

				rr := lsFiles.FileInfoTeams{
					SeasonID:      seasonID,
					CompetitionID: competitionID,
					Count:         dat[0],
					Name:          v.Name(),
				}

				fileList = append(fileList, rr)
			}

		} else {
			log.Printf("Wrong data format : %s", v.Name())
		}

	}

	return fileList, nil
}
