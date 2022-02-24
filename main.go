package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
)

// const exampleFile = "a_an_example"

type Skill struct {
	Name  string
	Level int
}

type Project struct {
	Name             string
	Duration         int
	BestBefore       int
	CompletionPoints int
	SkillsRequired   []Skill
	StartDay         int
	Contributors     []string
}

type Contributor struct {
	Name        string
	Skills      map[string]Skill
	AvailableOn int
}

func (c *Contributor) LevelIn(skillName string) int {
	skill, ok := c.Skills[skillName]
	if ok {
		return skill.Level
	}
	return 0
}

func (c *Contributor) AvailableFor(project Project) bool {
	return c.AvailableOn+project.Duration <= project.BestBefore
}

func (c *Contributor) IsPartOf(contributors []string) bool {
	for _, projectContributor := range contributors {
		if c.Name == projectContributor {
			return true
		}
	}
	return false
}

type ProjectWithContributors struct {
	ProjectName  string
	Contributors []string
}

func main() {
	files, err := ioutil.ReadDir("inputs")
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".txt") {
			example := strings.Split(file.Name(), ".")[0]
			parseFile(example)
		}
	}
}

func parseFile(exampleFile string) {
	content, err := ioutil.ReadFile(fmt.Sprintf("inputs/%s.in.txt", exampleFile))
	if err != nil {
		log.Fatal(err)
	}

	contributors := make(map[string]Contributor)
	//contributorsBySkill := make(map[string]Project)

	scanner := bufio.NewScanner(strings.NewReader(string(content)))

	// Parse first line
	scanner.Scan()
	firstLine := scanner.Text()
	contributorsCount, projectsCount := parseLineWith2int(firstLine)

	projectsWithContributors := []ProjectWithContributors{}

	for i := 0; i < contributorsCount; i++ {
		scanner.Scan()
		contributorName, skillsCount := parseLineWith1String1Int(scanner.Text())

		skills := make(map[string]Skill)
		for j := 0; j < skillsCount; j++ {
			scanner.Scan()
			skillName, skillLevel := parseLineWith1String1Int(scanner.Text())

			skill := Skill{
				Name:  skillName,
				Level: skillLevel,
			}
			skills[skillName] = skill
		}

		contributors[contributorName] = Contributor{
			Name:        contributorName,
			Skills:      skills,
			AvailableOn: 0,
		}
	}

	projects := scanProjects(projectsCount, scanner)

	sort.Slice(projects, func(i, j int) bool {
		return projects[i].Duration < projects[j].Duration
		//return projects[i].BestBefore < projects[j].BestBefore
		//return projects[i].CompletionPoints > projects[j].CompletionPoints
	})

	var projects2 []Project

	for _, project := range projects {
		err := getProjectWithContributors(&project, contributors)
		if err != nil {
			fmt.Println("Skipped: " + project.Name)
			continue
		}

		resultingProjectWithContributor := ProjectWithContributors{
			ProjectName:  project.Name,
			Contributors: project.Contributors,
		}
		projectsWithContributors = append(projectsWithContributors, resultingProjectWithContributor)
		projects2 = append(projects2, project)
	}

	// Write to result file
	result := fmt.Sprintf("%d", len(projectsWithContributors))
	for _, project := range projectsWithContributors {
		result += fmt.Sprintf("\n%s\n%s", project.ProjectName, strings.Join(project.Contributors, " "))
	}
	err = os.WriteFile(fmt.Sprintf("outputs/%s.out.txt", exampleFile), []byte(result), 0644)
	if err != nil {
		log.Fatal(err)
	}

	// Write to log file
	contributorsJson, err := json.MarshalIndent(contributors, "", "\t")
	checkErr(err)
	projectsJson, err := json.MarshalIndent(projects2, "", "\t")
	checkErr(err)
	projectsWithContributorsJson, err := json.MarshalIndent(projectsWithContributors, "", "\t")
	checkErr(err)
	err = os.WriteFile(fmt.Sprintf("logs/%s.contributors", exampleFile), []byte(fmt.Sprintf("%v", string(contributorsJson))), 0644)
	err = os.WriteFile(fmt.Sprintf("logs/%s.projects", exampleFile), []byte(fmt.Sprintf("%v", string(projectsJson))), 0644)
	err = os.WriteFile(fmt.Sprintf("logs/%s.result", exampleFile), []byte(fmt.Sprintf("%v", string(projectsWithContributorsJson))), 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func getProjectWithContributors(project *Project, contributors map[string]Contributor) error {
	projectContributors := []string{}
	skillUsedByContributors := make(map[string]Skill)
	for _, requiredSkill := range project.SkillsRequired {
		isMentorAvailable := IsMentorAvailableForProject(requiredSkill, contributors, projectContributors)
		contributorName := getContributorForSkill(*project, contributors, projectContributors, requiredSkill, isMentorAvailable)
		if contributorName == "" {
			return fmt.Errorf("skipped")
		}
		skillUsedByContributors[contributorName] = requiredSkill
		projectContributors = append(projectContributors, contributorName)
	}

	project.StartDay = 0
	for _, projectContributor := range projectContributors {
		contributor := contributors[projectContributor]
		if contributor.AvailableOn > project.StartDay {
			project.StartDay = contributor.AvailableOn
		}
	}

	for _, projectContributor := range projectContributors {
		contributor := contributors[projectContributor]
		skillUsedByContributor := skillUsedByContributors[contributor.Name]
		contributorSkillLevel := contributor.LevelIn(skillUsedByContributor.Name)
		if contributorSkillLevel <= skillUsedByContributor.Level {
			contributorSkill := contributors[projectContributor].Skills[skillUsedByContributor.Name]
			contributorSkill.Level = contributorSkill.Level + 1
			contributor.Skills[skillUsedByContributor.Name] = contributorSkill
		}
		contributor.AvailableOn = project.StartDay + project.Duration
		contributors[projectContributor] = contributor
	}

	project.Contributors = projectContributors
	return nil
}

func getContributorForSkill(project Project, contributors map[string]Contributor, projectContributors []string, requiredSkill Skill, mentorAvailable bool) string {
	// First iteration: Look for a contributor available before the end of project
	bestContributor := ""
	//lowerConstributorLevel := 99999999999
	lowerStartDay := 99999999999
	for _, contributor := range contributors {
		contributorSkillLevel := contributor.LevelIn(requiredSkill.Name)
		if contributorSkillLevel < requiredSkill.Level-1 {
			continue
		}

		if contributorSkillLevel == requiredSkill.Level-1 && !mentorAvailable {
			continue
		}

		if contributor.IsPartOf(projectContributors) {
			continue
		}

		if !contributor.AvailableFor(project) {
			continue
		}

		//if contributorSkillLevel < lowerConstributorLevel {
		//	bestContributor = contributor.Name
		//	lowerConstributorLevel = contributorSkillLevel
		//}
		if contributor.AvailableOn < lowerStartDay {
			bestContributor = contributor.Name
			lowerStartDay = contributor.AvailableOn
		}
	}

	if bestContributor != "" {
		return bestContributor
	}

	//// Second iteration: Look for a contributor even if not available
	////lowerConstributorLevel = 99999999999
	//for _, contributor := range contributors {
	//	contributorSkillLevel := contributor.LevelIn(requiredSkill.Name)
	//	if contributorSkillLevel < requiredSkill.Level {
	//		continue
	//	}
	//
	//	if contributor.IsPartOf(projectContributors) {
	//		continue
	//	}
	//
	//	//if contributorSkillLevel < lowerConstributorLevel {
	//	//	bestContributor = contributor.Name
	//	//	lowerConstributorLevel = contributorSkillLevel
	//	//}
	//	if contributor.AvailableOn < lowerStartDay {
	//		bestContributor = contributor.Name
	//		lowerStartDay = contributor.AvailableOn
	//	}
	//}
	return bestContributor
}

func getContributorsWithExactlevel(projectSkill Skill, contributors map[string]Contributor) []Contributor {
	potentialMentors := []Contributor{}
	for _, contributor := range contributors {
		if contributor.LevelIn(projectSkill.Name) == projectSkill.Level {
			potentialMentors = append(potentialMentors, contributor)
		}
	}
	return potentialMentors
}

func getPotentialMentors(projectSkill Skill, contributors map[string]Contributor) []Contributor {
	potentialMentors := []Contributor{}
	for _, contributor := range contributors {
		if contributor.LevelIn(projectSkill.Name) >= projectSkill.Level {
			potentialMentors = append(potentialMentors, contributor)
		}
	}
	return potentialMentors
}

func IsMentorAvailableForProject(projectSkill Skill, contributorsMap map[string]Contributor, projectContributors []string) bool {
	for _, contributor := range projectContributors {
		contributor := contributorsMap[contributor]
		if contributor.LevelIn(projectSkill.Name) >= projectSkill.Level {
			return true
		}
	}
	return false
}

func getPotentialMentees(projectSkill Skill, contributors map[string]Contributor) []Contributor {
	potentialMentees := []Contributor{}
	for _, contributor := range contributors {
		if contributor.LevelIn(projectSkill.Name) == projectSkill.Level-1 {
			potentialMentees = append(potentialMentees, contributor)
		}
	}
	return potentialMentees
}

func scanProjects(projectsCount int, scanner *bufio.Scanner) []Project {
	projects := []Project{}
	for i := 0; i < projectsCount; i++ {
		scanner.Scan()
		line := scanner.Text()
		lineSplitted := strings.Split(line, " ")

		projects = append(projects)

		skillsCount := Atoi(lineSplitted[4])
		var skills []Skill
		for j := 0; j < skillsCount; j++ {
			scanner.Scan()
			skillName, skillLevel := parseLineWith1String1Int(scanner.Text())

			skill := Skill{
				Name:  skillName,
				Level: skillLevel,
			}
			skills = append(skills, skill)
		}

		// Add new project to the list of projects
		projects = append(projects, Project{
			Name:             lineSplitted[0],
			Duration:         Atoi(lineSplitted[1]),
			BestBefore:       Atoi(lineSplitted[3]),
			CompletionPoints: Atoi(lineSplitted[2]),
			SkillsRequired:   skills,
		})
	}
	return projects
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func Atoi(s string) int {
	i, err := strconv.Atoi(s)
	checkErr(err)
	return i
}

func parseLineWith1String1Int(line string) (string, int) {
	lineSplitted := strings.Split(line, " ")
	return lineSplitted[0], Atoi(lineSplitted[1])
}

func parseLineWith2int(line string) (int, int) {
	lineSplitted := strings.Split(line, " ")
	return Atoi(lineSplitted[0]), Atoi(lineSplitted[1])
}

func RemoveIndex(s []string, index int) []string {
	return append(s[:index], s[index+1:]...)
}
